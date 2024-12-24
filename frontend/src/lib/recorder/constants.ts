export const noiseCaptureConfig = {
  min_decibels: -45, // Noise detection sensitivity
  max_blank_time: 700, // Maximum time to consider a blank (ms)
};

export const SAMPLE_RATE = 32000;
export const AUDIO_NUM_CHANNELS = 1;
export const EXPORT_MIME_TYPE = "audio/wav";

export const audio_options = {
  sampleRate: SAMPLE_RATE,
};

export const mediaStreamConstraints: MediaStreamConstraints = {
  audio: {
    echoCancellation: true,
    noiseSuppression: true,
    channelCount: AUDIO_NUM_CHANNELS,
    sampleRate: SAMPLE_RATE,
  },
  video: false,
};

export const defaultMicrophoneConfig = {
  bufferLen: 4096,
  numChannels: AUDIO_NUM_CHANNELS,
  mimeType: EXPORT_MIME_TYPE,
};
