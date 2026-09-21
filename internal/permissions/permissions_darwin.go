// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

//go:build darwin

package permissions

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework AVFoundation
#import <Cocoa/Cocoa.h>
#import <AVFoundation/AVFoundation.h>

static int microphoneStatus(void) {
    switch ([AVCaptureDevice authorizationStatusForMediaType:AVMediaTypeAudio]) {
    case AVAuthorizationStatusAuthorized: return 1;
    case AVAuthorizationStatusDenied: return 2;
    case AVAuthorizationStatusRestricted: return 3;
    default: return 0;
    }
}

// Shows the system prompt the first time and waits for the answer. Once
// decided, it returns at once. Safe to block: the handler runs off the main
// thread, and Wails calls bindings from goroutines.
static int requestMicrophone(void) {
    dispatch_semaphore_t done = dispatch_semaphore_create(0);
    __block BOOL granted = NO;
    [AVCaptureDevice requestAccessForMediaType:AVMediaTypeAudio completionHandler:^(BOOL ok) {
        granted = ok;
        dispatch_semaphore_signal(done);
    }];
    dispatch_semaphore_wait(done, DISPATCH_TIME_FOREVER);
    return granted ? 1 : microphoneStatus();
}

static void openMicrophonePreferences(void) {
    NSURL *url = [NSURL URLWithString:@"x-apple.systempreferences:com.apple.preference.security?Privacy_Microphone"];
    [[NSWorkspace sharedWorkspace] openURL:url];
}
*/
import "C"

func RequestMicrophone() MicrophoneStatus {
	switch C.requestMicrophone() {
	case 1:
		return MicrophoneGranted
	case 2:
		return MicrophoneDenied
	case 3:
		return MicrophoneRestricted
	default:
		return MicrophoneUndetermined
	}
}

func OpenMicrophoneSettings() {
	C.openMicrophonePreferences()
}
