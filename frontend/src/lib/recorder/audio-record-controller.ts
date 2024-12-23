import Recorder from "./recorder";
import SequentializerStatus from "./sequentializer-status";

const noiseCaptureConfig = {
  min_decibels: -40, // Noise detection sensitivity
  max_blank_time: 500, // Maximum time to consider a blank (ms)
};

const SAMPLE_RATE = 16000;
const AUDIO_NUM_CHANNELS = 1;
const EXPORT_MIME_TYPE = "audio/wav";

const audio_options = {
  sampleRate: SAMPLE_RATE,
};

const mediaStreamConstraints: MediaStreamConstraints = {
  audio: {
    echoCancellation: true,
    noiseSuppression: true,
    channelCount: AUDIO_NUM_CHANNELS,
    sampleRate: SAMPLE_RATE,
  },
  video: false,
};

type SequentializeCallback = (data: Blob) => void;

export class AudioRecordController {
  private static _instance: AudioRecordController;

  private audioContext: AudioContext;

  private sequentializerAnalyser?: {
    analyser: AnalyserNode;
    data: Uint8Array;
  };

  private sequentializerStatus?: SequentializerStatus;

  private sequentializing: boolean = false;

  private _onSequentializeCallbacks: SequentializeCallback[] = [];

  private mediaStream?: MediaStream;

  private recorder: Recorder;

  private constructor() {
    this.audioContext = new AudioContext(audio_options);

    this.recorder = new Recorder(this.audioContext);

    // This is a workaround for a bug in Safari.
    // See: https://bugs.webkit.org/show_bug.cgi?id=153693
    if (this.audioContext.state === "suspended") {
      this.audioContext.resume();
    }
  }

  static async create() {
    if (this._instance) {
      throw new Error("RecorderController.init() must be called only once.");
    }

    this._instance = new AudioRecordController();

    return this._instance;
  }

  private initSequentializer(mediaStream: MediaStream) {
    const analyser = this.audioContext.createAnalyser();
    const streamNode = this.audioContext.createMediaStreamSource(mediaStream);
    streamNode.connect(analyser);
    analyser.minDecibels = noiseCaptureConfig.min_decibels;

    this.sequentializerAnalyser = {
      analyser,
      data: new Uint8Array(analyser.frequencyBinCount),
    };

    this.sequentializerStatus = new SequentializerStatus();
  }

  private async startSequentializer() {
    this.sequentializing = true;

    const { analyser } = this.sequentializerAnalyser!;
    const data = new Uint8Array(analyser.frequencyBinCount);

    let silenceStart = Date.now();
    let triggered = false;

    const loop = async (time: number) => {
      if (!this.sequentializing || !this.recorder?.recording) {
        return;
      }

      requestAnimationFrame(loop);

      analyser.getByteFrequencyData(data);

      if (data.some((v) => v)) {
        if (triggered) {
          triggered = false;
          // noise detected
        }
        silenceStart = time;
      }

      if (
        !triggered &&
        time - silenceStart > noiseCaptureConfig.max_blank_time
      ) {
        // no noise detected
        triggered = true;
        if (this._onSequentializeCallbacks.length > 0) {
          // Enable pending the exportation
          this.sequentializerStatus?.enableInPending();

          const audioData = this.recorder.export();
          // this.sequentializeBlobs.push(audioData);

          // disable pending the exportation
          this.sequentializerStatus?.disableInPending();
          // clear all buffer from the recorder
          this.recorder.clear();

          this._onSequentializeCallbacks.forEach((callback) => {
            callback(audioData);
          });
        }
      }
    };

    requestAnimationFrame(loop);
  }

  private async stopSequentializer() {
    this.sequentializing = false;
  }

  private async initUserAgent() {
    if (this.mediaStream) {
      return;
    }

    this.mediaStream ??= await navigator.mediaDevices.getUserMedia(
      mediaStreamConstraints
    );

    await this.audioContext.resume();

    await this.recorder.init(this.mediaStream);

    this.initSequentializer(this.mediaStream);
  }

  /**
   * Callback called only when a noise is detected or when the recording stopped,
   */
  public onSequentialize(callback: SequentializeCallback) {
    this._onSequentializeCallbacks.push(callback);

    return () => {
      this._onSequentializeCallbacks = this._onSequentializeCallbacks.filter(
        (cb) => cb !== callback
      );
    };
  }

  public async startRecording() {
    await this.initUserAgent();
    await this.audioContext.resume();

    // start the sequentializer
    this.startSequentializer();
    // start the recorder
    await this.recorder.start();
  }

  public async stopRecording() {
    if (!this.recorder.recording) {
      return;
    }

    const blob = await this.recorder.stop();
    // stop the sequentializer
    this.stopSequentializer();

    this._onSequentializeCallbacks.forEach((callback) => {
      callback(blob);
    });
  }
}
