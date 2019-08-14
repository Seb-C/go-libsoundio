/*
 * Copyright (c) 2019 Zenichi Amano
 *
 * This file is part of libsoundio, which is MIT licensed.
 * See http://opensource.org/licenses/MIT
 */

// Package soundio is as set of bindings for the libsoundio sound library.
package soundio

/*
#cgo LDFLAGS: -lsoundio -lm
#include <soundio/soundio.h>
#include <stdlib.h>

extern void soundioOnDevicesChange(struct SoundIo *);
extern void soundioOnBackendDisconnect(struct SoundIo *, int);
extern void soundioOnEventsSignal(struct SoundIo *);

static void setSoundIoCallback(struct SoundIo *io) {
	io->on_devices_change = soundioOnDevicesChange;
	io->on_backend_disconnect = soundioOnBackendDisconnect;
	io->on_events_signal = soundioOnEventsSignal;
}
*/
import "C"
import "unsafe"

const (
	MaxChannels int = C.SOUNDIO_MAX_CHANNELS
)

type SoundIo struct {
	ptr                 *C.struct_SoundIo
	appNamePtr          *C.char
	onDevicesChange     func(io *SoundIo)
	onBackendDisconnect func(io *SoundIo, err error)
	onEventsSignal      func(io *SoundIo)
}

//export soundioOnDevicesChange
func soundioOnDevicesChange(nativeIo *C.struct_SoundIo) {
	io := (*SoundIo)(nativeIo.userdata)
	if io.onDevicesChange != nil {
		io.onDevicesChange(io)
	}
}

//export soundioOnBackendDisconnect
func soundioOnBackendDisconnect(nativeIo *C.struct_SoundIo, err C.int) {
	io := (*SoundIo)(nativeIo.userdata)
	if io.onBackendDisconnect != nil {
		io.onBackendDisconnect(io, convertToError(err))
	}
}

//export soundioOnEventsSignal
func soundioOnEventsSignal(nativeIo *C.struct_SoundIo) {
	io := (*SoundIo)(nativeIo.userdata)
	if io.onEventsSignal != nil {
		io.onEventsSignal(io)
	}
}

// fields

// GetCurrentBackend returns current backend.
func (s *SoundIo) GetCurrentBackend() Backend {
	return Backend(int(s.ptr.current_backend))
}

// GetAppName returns application name.
func (s *SoundIo) GetAppName() string {
	return C.GoString(s.ptr.app_name)
}

// SetAppName sets application name.
// PulseAudio uses this for "application name".
// JACK uses this for `client_name`.
// Must not contain a colon (":").
func (s *SoundIo) SetAppName(name string) {
	if s.appNamePtr != nil {
		C.free(unsafe.Pointer(s.appNamePtr))
	}
	s.ptr.app_name = C.CString(name)
}

// functions

// Version returns the version number string of libsoundio.
func Version() string {
	return C.GoString(C.soundio_version_string())
}

// VersionMajor returns the major version number of libsoundio.
func VersionMajor() int {
	return int(C.soundio_version_major())
}

// VersionMinor returns the minor version number of libsoundio.
func VersionMinor() int {
	return int(C.soundio_version_minor())
}

// VersionPatch returns the patch version number of libsoundio.
func VersionPatch() int {
	return int(C.soundio_version_patch())
}

// GetBytesPerSample returns bytes per sample.
// Returns -1 on invalid format.
func GetBytesPerSample(format Format) int {
	return int(C.soundio_get_bytes_per_sample(uint32(format)))
}

// GetBytesPerFrame returns bytes per frame.
// A frame is one sample per channel.
func GetBytesPerFrame(format Format, channelCount int) int {
	return int(C.soundio_get_bytes_per_frame(uint32(format), C.int(channelCount)))
}

// GetBytesPerSecond returns bytes per second.
// Sample rate is the number of frames per second.
func GetBytesPerSecond(format Format, channelCount int, sampleRate int) int {
	return int(C.soundio_get_bytes_per_second(uint32(format), C.int(channelCount), C.int(sampleRate)))
}

// Create a SoundIo context. You may create multiple instances of this to connect to multiple backends. Sets all fields to defaults.
func Create() *SoundIo {
	ptr := C.soundio_create()
	io := &SoundIo{
		ptr: ptr,
	}
	ptr.userdata = unsafe.Pointer(io)
	C.setSoundIoCallback(ptr)
	return io
}

// Destroy releases resources.
func (s *SoundIo) Destroy() {
	if s.appNamePtr != nil {
		C.free(unsafe.Pointer(s.appNamePtr))
	}
	s.appNamePtr = nil

	if s.ptr != nil {
		s.ptr.userdata = nil
		C.soundio_destroy(s.ptr)
	}
	s.ptr = nil
}

// Connect tries to connect on all available backends in order.
func (s *SoundIo) Connect() error {
	return convertToError(C.soundio_connect(s.ptr))
}

// ConnectBackend connect to backend.
// Instead of calling Connect function you may call this function to try a specific backend.
func (s *SoundIo) ConnectBackend(backend Backend) error {
	return convertToError(C.soundio_connect_backend(s.ptr, uint32(backend)))
}

// Disconnect disconnect from backend.
func (s *SoundIo) Disconnect() {
	C.soundio_disconnect(s.ptr)
}

// BackendCount returns the number of available backends.
func (s *SoundIo) BackendCount() int {
	return int(C.soundio_backend_count(s.ptr))
}

// GetBackend returns the available backend at the specified index (0 <= index < BackendCount)
func (s *SoundIo) GetBackend(index int) Backend {
	return Backend(C.soundio_get_backend(s.ptr, C.int(index)))
}

// FlushEvents atomically updates information for all connected devices.
func (s *SoundIo) FlushEvents() {
	C.soundio_flush_events(s.ptr)
}

// WaitEvents calls FlushEvents then blocks until another event
// is ready or you call Wakeup. Be ready for spurious wakeups.
func (s *SoundIo) WaitEvents() {
	C.soundio_wait_events(s.ptr)
}

// Wakeup makes WaitEvents stop blocking.
func (s *SoundIo) Wakeup() {
	C.soundio_wakeup(s.ptr)
}

// ForceDeviceScan rescan device If necessary.
func (s *SoundIo) ForceDeviceScan() {
	C.soundio_force_device_scan(s.ptr)
}

// InputDeviceCount returns the number of input devices.
// Returns -1 if you never called FlushEvents.
func (s *SoundIo) InputDeviceCount() int {
	return int(C.soundio_input_device_count(s.ptr))
}

// OutputDeviceCount returns the number of output devices.
// Returns -1 if you never called FlushEvents.
func (s *SoundIo) OutputDeviceCount() int {
	return int(C.soundio_output_device_count(s.ptr))
}

// GetInputDevice returns a device.
// Call RemoveReference when done.
// `index` must be 0 <= index < InputDeviceCount.
func (s *SoundIo) GetInputDevice(index int) *Device {
	return &Device{
		ptr: C.soundio_get_input_device(s.ptr, C.int(index)),
	}
}

// GetOutputDevice returns a device.
// Call RemoveReference when done.
// `index` must be 0 <= index < OutputDeviceCount
func (s *SoundIo) GetOutputDevice(index int) *Device {
	return &Device{
		ptr: C.soundio_get_output_device(s.ptr, C.int(index)),
	}
}

// DefaultInputDeviceIndex returns the index of the default input device
// returns -1 if there are no devices or if you never called FlushEvents.
func (s *SoundIo) DefaultInputDeviceIndex() int {
	return int(C.soundio_default_input_device_index(s.ptr))
}

// DefaultOutputDeviceIndex returns the index of the default output device
// returns -1 if there are no devices or if you never called FlushEvents.
func (s *SoundIo) DefaultOutputDeviceIndex() int {
	return int(C.soundio_default_output_device_index(s.ptr))
}

// RingBufferCreate creates a ring buffer is a single-reader single-writer lock-free fixed-size queue.
func (s *SoundIo) RingBufferCreate(requestedCapacity int) *RingBuffer {
	return &RingBuffer{
		ptr: C.soundio_ring_buffer_create(s.ptr, C.int(requestedCapacity)),
	}
}
