import {
  IMediaRecorder,
  MediaRecorder,
  register,
} from "extendable-media-recorder";
import { connect } from "extendable-media-recorder-wav-encoder";

export default class Recorder {
  private mediaRecorder?: IMediaRecorder;

  private chunks: Blob[] = [];

  private _recording: boolean = false;

  constructor(private audioContext: AudioContext) {}

  public async init(mediaStream: MediaStream) {
    await register(await connect());

    const mediaStreamAudioSourceNode = new MediaStreamAudioSourceNode(
      this.audioContext,
      { mediaStream }
    );
    const mediaStreamAudioDestinationNode = new MediaStreamAudioDestinationNode(
      this.audioContext
    );

    mediaStreamAudioSourceNode.connect(mediaStreamAudioDestinationNode);

    this.mediaRecorder = new MediaRecorder(
      mediaStreamAudioDestinationNode.stream,
      { mimeType: "audio/wav" }
    );
  }

  public get recording() {
    return this._recording;
  }

  private mustBeInitialized() {
    if (!this.mediaRecorder) {
      throw new Error("MediaRecorder is not initialized");
    }
  }

  public start() {
    this.mustBeInitialized();

    this.mediaRecorder!.onstart = () => {
      this._recording = true;
    };

    this.mediaRecorder!.onstop = () => {
      this._recording = false;
    };

    this.mediaRecorder!.ondataavailable = (event) => {
      this.chunks.push(event.data);
    };

    this.mediaRecorder!.start();
  }

  public stop() {
    this.mustBeInitialized();
    this.mediaRecorder!.stop();
    this.mediaRecorder!.ondataavailable = null;
    this.mediaRecorder!.onstop = null;
    this.mediaRecorder!.onstart = null;

    const chunks = this.chunks.slice();
    this.chunks = [];

    return new Blob(chunks);
  }

  public export() {
    return new Blob(this.chunks);
  }

  public clear() {
    this.chunks = [];
  }
}
