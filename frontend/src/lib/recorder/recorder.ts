import Microphone from "./microphone";

const defaultConfig = {
  nFrequencyBars: 255,
  onAnalysed: null,
};

interface RecorderResult {
  blob: Blob;
  lastSequence?: Blob;
  buffer: Float32Array[];
  sampleRate?: number;
}

export default class Recorder {
  config: { nFrequencyBars: number; onAnalysed: any };
  static download: (blob: Blob, filename?: string) => void;
  audioContext: AudioContext;
  audioInput: MediaStreamAudioSourceNode | null;
  realAudioInput: MediaStreamAudioSourceNode | null;
  inputPoint: GainNode | null;
  audioRecorder?: Microphone;
  rafID: any;
  analyserContext: any;
  recIndex: number;
  stream: MediaStream | null;
  analyserNode: AnalyserNode | undefined;
  private cancelRequestAnimationFrame: number | undefined;

  constructor(
    audioContext: AudioContext,
    config?: Omit<typeof defaultConfig, "onAnalysed">
  ) {
    this.config = Object.assign({}, defaultConfig, config || {});

    this.audioContext = audioContext;
    this.audioInput = null;
    this.realAudioInput = null;
    this.inputPoint = null;
    // @ts-ignore
    this.audioRecorder = null;
    this.rafID = null;
    this.analyserContext = null;
    this.recIndex = 0;
    this.stream = null;

    this.updateAnalysers = this.updateAnalysers.bind(this);
  }

  init(stream: MediaStream) {
    return new Promise<void>((resolve) => {
      this.inputPoint = this.audioContext.createGain();

      this.stream = stream;

      this.realAudioInput = this.audioContext.createMediaStreamSource(stream);
      this.audioInput = this.realAudioInput;
      this.audioInput?.connect(this.inputPoint);

      this.analyserNode = this.audioContext.createAnalyser();
      this.analyserNode.fftSize = 2048;
      this.inputPoint.connect(this.analyserNode);

      this.audioRecorder = new Microphone(this.inputPoint);

      const zeroGain = this.audioContext.createGain();
      zeroGain.gain.value = 0.0;

      this.inputPoint.connect(zeroGain);
      zeroGain.connect(this.audioContext.destination);

      this.updateAnalysers();

      resolve();
    });
  }

  start() {
    return new Promise<MediaStream>((resolve, reject) => {
      if (!this.audioRecorder) {
        reject("Not currently recording");
        return;
      }

      this.audioRecorder.clear();
      this.audioRecorder.record();

      resolve(this.stream!);
    });
  }

  export(): Promise<RecorderResult> {
    return new Promise((resolve) => {
      this.audioRecorder?.getBuffer(
        (buffer: Float32Array[], sampleRate: number) => {
          this.audioRecorder?.exportWAV((blob: Blob) => {
            resolve({ buffer, blob, sampleRate });
          });
        }
      );
    });
  }

  stop() {
    return new Promise<RecorderResult>((resolve) => {
      this.audioRecorder?.stop();

      this.audioRecorder?.getBuffer(
        (buffer: Float32Array[], sampleRate: number) => {
          this.audioRecorder?.exportWAV((blob: Blob) =>
            resolve({ buffer, blob, sampleRate })
          );
        }
      );
    });
  }

  updateAnalysers() {
    if (this.config.onAnalysed) {
      this.cancelRequestAnimationFrame = requestAnimationFrame(
        this.updateAnalysers
      );

      const freqByteData = new Uint8Array(
        this.analyserNode?.frequencyBinCount!
      );

      this.analyserNode?.getByteFrequencyData(freqByteData);

      const data = new Array(255);
      let lastNonZero = 0;
      let datum;

      for (let idx = 0; idx < 255; idx += 1) {
        datum =
          Math.floor(freqByteData[idx]) - (Math.floor(freqByteData[idx]) % 5);

        if (datum !== 0) {
          lastNonZero = idx;
        }

        data[idx] = datum;
      }

      this.config.onAnalysed({ data, lineTo: lastNonZero });
    } else if (this.cancelRequestAnimationFrame && this.config.onAnalysed) {
      cancelAnimationFrame(this.cancelRequestAnimationFrame);
    }
  }

  setOnAnalysed(handler: () => void) {
    this.config.onAnalysed = handler;
  }
}
