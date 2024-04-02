package streams

import (
	"errors"

	"github.com/dop251/goja"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/js/promises"
	"gopkg.in/guregu/null.v3"
)

// ReadableStream is a concrete instance of the general [readable stream] concept.
//
// It is adaptable to any chunk type, and maintains an internal queue to keep track of
// data supplied by the underlying source but not yet read by any consumer.
//
// [readable stream]: https://streams.spec.whatwg.org/#rs-class
type ReadableStream struct {
	// FIXME: should be a public property of the object exposed to the runtime instead
	// locked indicate whether or not the readable stream is locked to a reader
	locked bool

	// controller holds a [ReadableStreamDefaultController] or [ReadableByteStreamController] created
	// with the ability to control the state and queue of this stream.
	controller ReadableStreamController

	// detached is a boolean flag set to true when the stream is transferred
	detached bool

	// disturbed is true when the stream has been read from or canceled
	disturbed bool

	// reader holds the current reader of the stream if the stream is locked to a reader
	// or nil otherwise.
	reader any

	// state holds the current state of the stream
	state ReadableStreamState

	// storedError holds the error that caused the stream to be errored
	storedError any

	// FIXME: Do we really need it? Some WPTs manipulate this.
	Source any

	runtime *goja.Runtime
	vu      modules.VU
}

// Locked returns whether or not the readable stream is locked to a reader
// FIXME: this should be a property
func (stream *ReadableStream) Locked() bool {
	return stream.isLocked()
}

// Cancel cancels the stream and returns a Promise to the user
//
// FIXME: clarify the proper type to use for `reason` taking inspiration from
// what we already do in k6
func (stream *ReadableStream) Cancel(reason goja.Value) *goja.Promise {
	// 1. IsReadableStreamLocked(this) is true, return a promise rejected with a TypeError exception.
	if stream.isLocked() {
		promise, _, reject := promises.New(stream.vu)

		go func() {
			reject(newError(TypeError, "cannot cancel a locked stream"))
		}()

		return promise
	}

	// 2. Return ! ReadableStreamCancel(reason)
	// FIXME: align the `reason` type to make it consistent with k6
	return stream.cancel(reason)
}

// GetReader implements the [getReader] operation.
//
// [getReader]: https://streams.spec.whatwg.org/#rs-get-reader
func (stream *ReadableStream) GetReader(options *goja.Object) goja.Value {
	// 1. If options["mode"] does not exist, return ? AcquireReadableStreamDefaultReader(this).
	if options == nil || common.IsNullish(options) || common.IsNullish(options.Get("mode")) {
		defaultReader := stream.acquireDefaultReader()
		defaultReaderObj, err := NewReadableStreamDefaultReaderObject(defaultReader)
		if err != nil {
			common.Throw(stream.runtime, err)
		}

		return defaultReaderObj
	}

	// 2. Assert: options["mode"] is "byob".
	if options.Get("mode").String() != ReaderTypeByob {
		common.Throw(stream.runtime, newError(TypeError, "options.mode is not 'byob'"))
	}

	// 3. Return ? AcquireReadableStreamBYOBReader(this).
	return stream.runtime.ToValue(stream.acquireBYOBReader())
}

// Tee implements the [tee] operation.
//
// [tee]: https://streams.spec.whatwg.org/#rs-tee
func (stream *ReadableStream) Tee() goja.Value {
	return nil
}

// ReadableStreamState represents the current state of a ReadableStream
type ReadableStreamState string

const (
	// ReadableStreamStateReadable indicates that the stream is readable, and that more data may be read from the stream.
	ReadableStreamStateReadable = "readable"

	// ReadableStreamStateClosed indicates that the stream is closed and cannot be read from.
	ReadableStreamStateClosed = "closed"

	// ReadableStreamStateErrored indicates that the stream has been aborted (errored).
	ReadableStreamStateErrored = "errored"
)

// ReadableStreamType represents the type of the ReadableStream
type ReadableStreamType = string

const (
	// ReadableStreamTypeBytes indicates that the stream is a byte stream.
	ReadableStreamTypeBytes = "bytes"
)

// isLocked implements the specification's [IsReadableStreamLocked()] abstract operation.
//
// [IsReadableStreamLocked()]: https://streams.spec.whatwg.org/#is-readable-stream-locked
// FIXME: This should be called when getting the stream.locked property
func (stream *ReadableStream) isLocked() bool {
	return stream.reader != nil
}

// initialize implements the specification's [InitializeReadableStream()] abstract operation.
//
// [InitializeReadableStream()]: https://streams.spec.whatwg.org/#initialize-readable-stream
func (stream *ReadableStream) initialize() {
	stream.state = ReadableStreamStateReadable
	stream.reader = nil
	stream.storedError = nil
	stream.disturbed = false
}

// setupReadableByteStreamControllerFromUnderlyingSource implements the [specification]'s
// SetUpReadableByteStreamControllerFromUnderlyingSource abstract operation.
//
// [specification]: https://streams.spec.whatwg.org/#set-up-readable-byte-stream-controller-from-underlying-source
// FIXME: Try to unify it with setupReadableStreamDefaultControllerFromUnderlyingSource.
func (stream *ReadableStream) setupReadableByteStreamControllerFromUnderlyingSource(
	_ goja.Value,
	underlyingSourceDict UnderlyingSource,
	highWaterMark float64,
) {
	// 1. Let controller be a new ReadableByteStreamController.
	controller := &ReadableByteStreamController{}

	// 2. Let startAlgorithm be an algorithm that returns undefined.
	var startAlgorithm UnderlyingSourceStartCallback = func(*goja.Object) goja.Value {
		return goja.Undefined()
	}

	// 3. Let pullAlgorithm be an algorithm that returns a promise resolved with undefined.
	var pullAlgorithm UnderlyingSourcePullCallback = func(*goja.Object) *goja.Promise {
		return newResolvedPromise(stream.vu, goja.Undefined())
	}

	// 4. Let cancelAlgorithm be an algorithm that returns a promise resolved with undefined.
	var cancelAlgorithm UnderlyingSourceCancelCallback = func(any) *goja.Promise {
		return newResolvedPromise(stream.vu, goja.Undefined())
	}

	// 5. If underlyingSourceDict["start"] exists, then set startAlgorithm to an algorithm
	// which returns the result of invoking underlyingSourceDict["start"] with argument
	// list « controller » and callback this value underlyingSource.
	if underlyingSourceDict.startSet {
		startAlgorithm = func(obj *goja.Object) (v goja.Value) {
			return underlyingSourceDict.Start(obj)
		}
	}

	// 6. If underlyingSourceDict["pull"] exists, then set pullAlgorithm to an algorithm
	// which returns the result of invoking underlyingSourceDict["pull"] with argument
	// list « controller » and callback this value underlyingSource.
	if underlyingSourceDict.pullSet {
		pullAlgorithm = func(obj *goja.Object) (p *goja.Promise) {
			if e := stream.runtime.Try(func() {
				p = underlyingSourceDict.Pull(obj)
			}); e != nil {
				return newRejectedPromise(stream.vu, e.Value())
			}

			if p == nil {
				p, _, _ = promises.New(stream.vu)
			}
			return p
		}
	}

	// 7. If underlyingSourceDict["cancel"] exists, then set cancelAlgorithm to an algorithm
	// which takes an argument reason and returns the result of invoking underlyingSourceDict["cancel"]
	// with argument list « reason » and callback this value underlyingSource.
	if underlyingSourceDict.cancelSet {
		cancelAlgorithm = func(reason any) *goja.Promise {
			return underlyingSourceDict.Cancel(reason)
		}
	}

	// 8. Let autoAllocateChunkSize be underlyingSourceDict["autoAllocateChunkSize"], if it exists, or undefined otherwise.
	autoAllocateChunkSize := underlyingSourceDict.AutoAllocateChunkSize

	// 9. If autoAllocateChunkSize is 0, then throw a TypeError exception.
	if autoAllocateChunkSize.Valid && autoAllocateChunkSize.Int64 == 0 {
		common.Throw(stream.vu.Runtime(), newError(TypeError, "underlyingSource.[[autoAllocateChunkSize]] must be > 0"))
	}

	// 10. Perform ? SetUpReadableByteStreamController(stream, controller, startAlgorithm, pullAlgorithm, cancelAlgorithm, highWaterMark, autoAllocateChunkSize).
	stream.setUpReadableByteStreamController(controller, startAlgorithm, pullAlgorithm, cancelAlgorithm, highWaterMark, autoAllocateChunkSize)
}

// setupReadableStreamDefaultControllerFromUnderlyingSource implements the [specification]'s
// SetUpReadableStreamDefaultController abstract operation.
//
// [specification]: https://streams.spec.whatwg.org/#set-up-readable-stream-default-controller-from-underlying-source
func (stream *ReadableStream) setupReadableStreamDefaultControllerFromUnderlyingSource(
	underlyingSource goja.Value,
	underlyingSourceDict UnderlyingSource,
	highWaterMark float64,
	sizeAlgorithm SizeAlgorithm,
) {
	// 1. Let controller be a new ReadableStreamDefaultController.
	controller := &ReadableStreamDefaultController{}

	// 2. Let startAlgorithm be an algorithm that returns undefined.
	var startAlgorithm UnderlyingSourceStartCallback = func(*goja.Object) goja.Value {
		return goja.Undefined()
	}

	// 3. Let pullAlgorithm be an algorithm that returns a promise resolved with undefined.
	var pullAlgorithm UnderlyingSourcePullCallback = func(*goja.Object) *goja.Promise {
		return newResolvedPromise(stream.vu, goja.Undefined())
	}

	// 4. Let cancelAlgorithm be an algorithm that returns a promise resolved with undefined.
	var cancelAlgorithm UnderlyingSourceCancelCallback = func(any) *goja.Promise {
		return newResolvedPromise(stream.vu, goja.Undefined())
	}

	// 5. If underlyingSourceDict["start"] exists, then set startAlgorithm to an algorithm
	// which returns the result of invoking underlyingSourceDict["start"] with argument
	// list « controller » and callback this value underlyingSource.
	if underlyingSourceDict.startSet {
		call, ok := goja.AssertFunction(stream.runtime.ToValue(underlyingSourceDict.Start))
		if !ok {
			common.Throw(stream.runtime, errors.New("underlyingSource.[[start]] must be a function"))
		}

		startAlgorithm = func(obj *goja.Object) (v goja.Value) {
			var err error
			v, err = call(underlyingSource, obj)
			// FIXME: Do something meaningful with error
			if err != nil {
				panic(err)
			}

			return v
		}
	}

	// 6. If underlyingSourceDict["pull"] exists, then set pullAlgorithm to an algorithm which
	// returns the result of invoking underlyingSourceDict["pull"] with argument list
	// « controller » and callback this value underlyingSource.
	if underlyingSourceDict.pullSet {
		pullAlgorithm = func(obj *goja.Object) (p *goja.Promise) {
			if e := stream.runtime.Try(func() {
				p = underlyingSourceDict.Pull(obj)
			}); e != nil {
				return newRejectedPromise(stream.vu, e.Value())
			}

			if p == nil {
				p, _, _ = promises.New(stream.vu)
			}
			return p
		}
	}

	// 7. If underlyingSourceDict["cancel"] exists, then set cancelAlgorithm to an algorithm which takes an argument
	// reason and returns the result of invoking underlyingSourceDict["cancel"] with argument list « reason » and
	// callback this value underlyingSource.
	if underlyingSourceDict.cancelSet {
		cancelAlgorithm = func(reason any) *goja.Promise {
			return underlyingSourceDict.Cancel(reason)
		}
	}

	// 8. Perform ? SetUpReadableStreamDefaultController(...)
	stream.setupDefaultController(controller, startAlgorithm, pullAlgorithm, cancelAlgorithm, highWaterMark, sizeAlgorithm)
}

// setupDefaultController implements the specification's [SetUpReadableStreamDefaultController()] abstract operation.
//
// [SetUpReadableStreamDefaultController()]: https://streams.spec.whatwg.org/#set-up-readable-stream-default-controller
func (stream *ReadableStream) setupDefaultController(
	controller *ReadableStreamDefaultController,
	startAlgorithm UnderlyingSourceStartCallback,
	pullAlgorithm UnderlyingSourcePullCallback,
	cancelAlgorithm UnderlyingSourceCancelCallback,
	highWaterMark float64,
	sizeAlgorithm SizeAlgorithm,
) {
	rt := stream.vu.Runtime()

	// 1. Assert: stream.[[controller]] is undefined.
	if stream.controller != nil {
		common.Throw(rt, newError(AssertionError, "stream.[[controller]] is not undefined"))
	}

	// 2. Set controller.[[stream]] to stream.
	controller.stream = stream

	// 3. Perform ! ResetQueue(controller).
	controller.resetQueue()

	// 4. Set controller.[[started]], controller.[[closeRequested]], controller.[[pullAgain]], and
	// controller.[[pulling]] to false.
	controller.started = false
	controller.closeRequested = false
	controller.pullAgain = false
	controller.pulling = false

	// 5. Set controller.[[strategySizeAlgorithm]] to sizeAlgorithm and controller.[[strategyHWM]] to highWaterMark.
	controller.strategySizeAlgorithm = sizeAlgorithm
	controller.strategyHWM = highWaterMark

	// 6. Set controller.[[pullAlgorithm]] to pullAlgorithm.
	controller.pullAlgorithm = pullAlgorithm

	// 7. Set controller.[[cancelAlgorithm]] to cancelAlgorithm.
	controller.cancelAlgorithm = cancelAlgorithm

	// 8. Set stream.[[controller]] to controller.
	stream.controller = controller

	// 9. Let startResult be the result of performing startAlgorithm. (This might throw an exception.)
	controllerObj, err := controller.toObject()
	if err != nil {
		common.Throw(controller.stream.vu.Runtime(), newError(RuntimeError, err.Error()))
	}
	startResult := startAlgorithm(controllerObj)

	// 10. Let startPromise be a promise with startResult.
	var startPromise *goja.Promise
	if common.IsNullish(startResult) {
		startPromise = newResolvedPromise(controller.stream.vu, startResult)
	} else if p, ok := startResult.Export().(*goja.Promise); ok {
		startPromise = p
	} else {
		startPromise = newResolvedPromise(controller.stream.vu, startResult)
	}

	_, err = promiseThen(stream.vu.Runtime(), startPromise,
		// 11. Upon fulfillment of startPromise,
		func(goja.Value) {
			// 11.1. Set controller.[[started]] to true.
			controller.started = true

			// 11.2. Assert: controller.[[pulling]] is false.
			if controller.pulling {
				common.Throw(stream.vu.Runtime(), newError(AssertionError, "controller `pulling` state is not false"))
			}

			// 11.3. Assert: controller.[[pullAgain]] is false.
			if controller.pullAgain {
				common.Throw(stream.vu.Runtime(), newError(AssertionError, "controller `pullAgain` state is not false"))
			}

			// 11.4. Perform ! ReadableStreamDefaultControllerCallPullIfNeeded(controller).
			controller.callPullIfNeeded()
		},

		// 12. Upon rejection of startPromise with reason r,
		func(err goja.Value) {
			controller.error(err)
		},
	)
	if err != nil {
		common.Throw(stream.vu.Runtime(), err)
	}
}

// setUpReadableByteStreamController implements the specification's [SetUpReadableByteStreamController()] abstract operation.
//
// [SetUpReadableByteStreamController()]: https://streams.spec.whatwg.org/#set-up-readable-byte-stream-controller
// FIXME: Try to unify it with setupDefaultController.
func (stream *ReadableStream) setUpReadableByteStreamController(
	controller *ReadableByteStreamController,
	startAlgorithm UnderlyingSourceStartCallback,
	pullAlgorithm UnderlyingSourcePullCallback,
	cancelAlgorithm UnderlyingSourceCancelCallback,
	highWaterMark float64,
	autoAllocateChunkSize null.Int,
) {
	// 1. Assert: stream.[[controller]] is undefined.
	if stream.controller != nil {
		common.Throw(stream.vu.Runtime(), newError(AssertionError, "stream.[[controller]] is not undefined"))
	}

	// 2. If autoAllocateChunkSize is not undefined,
	if autoAllocateChunkSize.Valid {
		// 2.1. Assert: ! IsInteger(autoAllocateChunkSize) is true.
		// 2.1. Assert: autoAllocateChunkSize is positive.
		if autoAllocateChunkSize.Int64 <= 0 {
			common.Throw(stream.vu.Runtime(), newError(AssertionError, "underlyingSource.[[autoAllocateChunkSize]] must be > 0"))
		}
	}

	// 3. Set controller.[[stream]] to stream.
	controller.stream = stream

	// 4. Set controller.[[pullAgain]] and controller.[[pulling]] to false.
	controller.pullAgain = false
	controller.pulling = false

	// 5. Set controller.[[byobRequest]] to null.
	controller.ByobRequest = nil

	// 6. Perform ! ResetQueue(controller).
	controller.resetQueue()

	// 7. Set controller.[[closeRequested]] and controller.[[started]] to false.
	controller.closeRequested = false
	controller.started = false

	// 8. Set controller.[[strategyHWM]] to highWaterMark.
	controller.strategyHWM = int64(highWaterMark)

	// 9. Set controller.[[pullAlgorithm]] to pullAlgorithm.

	// 10. Set controller.[[cancelAlgorithm]] to cancelAlgorithm.

	// 11. Set controller.[[autoAllocateChunkSize]] to autoAllocateChunkSize.
	if autoAllocateChunkSize.Valid {
		controller.autoAllocateChunkSize = autoAllocateChunkSize.Int64
	}

	// 12. Set controller.[[pendingPullIntos]] to a new empty list.
	controller.pendingPullIntos = make([]PullIntoDescriptor, 0)

	// 13. Set stream.[[controller]] to controller.
	stream.controller = controller

	// 14. Let startResult be the result of performing startAlgorithm.

	// 15. Let startPromise be a promise resolved with startResult.

	// 16. Upon fulfillment of startPromise,

	// 17. Set controller.[[started]] to true.

	// 18. Assert: controller.[[pulling]] is false.

	// 19. Assert: controller.[[pullAgain]] is false.

	// 20. Perform ! ReadableByteStreamControllerCallPullIfNeeded(controller).

	// 21. Upon rejection of startPromise with reason r,

	// 22. Perform ! ReadableByteStreamControllerError(controller, r).
}

// setupDefaultReader implements the specification's [SetUpReadableStreamDefaultReader()] abstract operation.
//
// [SetUpReadableStreamDefaultReader()]: https://streams.spec.whatwg.org/#set-up-readable-stream-default-reader
func (stream *ReadableStream) setupDefaultReader(reader *ReadableStreamDefaultReader) {
	// 1. If ! IsReadableStreamLocked(stream) is true, throw a TypeError exception.
	if stream.isLocked() {
		common.Throw(stream.vu.Runtime(), newError(TypeError, "cannot create a reader for a locked stream"))
	}

	// 2. Perform ! ReadableStreamReaderGenericInitialize(reader, stream).
	// TODO: we assume that the reader is a ReadableStreamDefaultReader, but we should probably
	// FIXME: sets stream to be a generic reader under the hood, and we really want it to be a default reader...
	// reader.ReadableStreamGenericReader.initialize(stream)
	ReadableStreamReaderGenericInitialize(reader, stream)

	// 3.
	reader.readRequests = []ReadRequest{}
}

// setupBYOBReader implements the specification's [SetUpReadableStreamBYOBReader()] abstract operation.
//
// [SetUpReadableStreamBYOBReader()]: https://streams.spec.whatwg.org/#set-up-readable-stream-byob-reader
func (stream *ReadableStream) setupBYOBReader(reader *ReadableStreamBYOBReader) {
	// 1.
	if stream.isLocked() {
		common.Throw(stream.vu.Runtime(), newError(TypeError, "cannot create a reader for a locked stream"))
	}

	// 2.
	_, ok := stream.controller.(*ReadableByteStreamController)
	if !ok {
		common.Throw(stream.vu.Runtime(), newError(TypeError, "stream controller is not a ReadableByteStreamController"))
	}

	// 3.
	// reader.ReadableStreamGenericReader.initialize(stream)
	ReadableStreamReaderGenericInitialize(reader, stream)

	// 4.
	reader.readIntoRequests = []ReadIntoRequest{}
}

// acquireDefaultReader implements the specification's [AcquireReadableStreamDefaultReader] algorithm.
//
// [AcquireReadableStreamDefaultReader]: https://streams.spec.whatwg.org/#acquire-readable-stream-reader
func (stream *ReadableStream) acquireDefaultReader() *ReadableStreamDefaultReader {
	// 1. Let reader be a new ReadableStreamDefaultReader.
	reader := &ReadableStreamDefaultReader{}

	// 2. Perform ? SetUpReadableStreamDefaultReader(reader, stream).
	reader.setup(stream)

	// 3. Return reader.
	return reader
}

// acquireBYOBReader implements the specification's [AcquireReadableStreamBYOBReader()] abstract operation.
//
// [AcquireReadableStreamBYOBReader()]: https://streams.spec.whatwg.org/#acquire-readable-stream-byob-reader
func (stream *ReadableStream) acquireBYOBReader() *ReadableStreamBYOBReader {
	// 1. let reader b a new ReadableStreamBYOBReader
	// FIXME: remove this?
	// reader := NewReadableStreamBYOBReader(stream)
	reader := &ReadableStreamBYOBReader{}

	// 2.
	// TODO: implement this!
	stream.setupBYOBReader(reader)

	// 3.
	return reader
}

// addReadRequest implements the specification's [ReadableStreamAddReadRequest()] abstract operation.
//
// [ReadableStreamAddReadRequest()]: https://streams.spec.whatwg.org/#readable-stream-add-read-request
func (stream *ReadableStream) addReadRequest(readRequest ReadRequest) {
	// 1.
	defaultReader, ok := stream.reader.(*ReadableStreamDefaultReader)
	if !ok {
		common.Throw(stream.vu.Runtime(), newError(AssertionError, "reader is not a ReadableStreamDefaultReader"))
	}

	// 2.
	if stream.state != ReadableStreamStateReadable {
		common.Throw(stream.vu.Runtime(), newError(AssertionError, "stream is not readable"))
	}

	// 3.
	defaultReader.readRequests = append(defaultReader.readRequests, readRequest)
}

// cancel implements the specification's [ReadableStreamCancel()] abstract operation.
//
// [ReadableStreamCancel()]: https://streams.spec.whatwg.org/#readable-stream-cancel
func (stream *ReadableStream) cancel(reason goja.Value) *goja.Promise {
	// 1.
	stream.disturbed = true

	// 2.
	if stream.state == ReadableStreamStateClosed {
		return newResolvedPromise(stream.vu, goja.Undefined())
	}

	// 3.
	if stream.state == ReadableStreamStateErrored {
		return newRejectedPromise(stream.vu, stream.storedError)
	}

	// 4.
	stream.close()

	// 5.
	reader := stream.reader

	// 6.
	byobReader, isBYOBReader := reader.(ReadableStreamBYOBReader)
	if reader != nil && isBYOBReader {
		// 6.1. Let readIntoRequests be reader.[[readIntoRequests]].
		readIntoRequests := byobReader.readIntoRequests

		// 6.2. Set reader.[[readIntoRequests]] to an empty list.
		byobReader.readIntoRequests = []ReadIntoRequest{}

		// 6.3. For each readIntoRequest of readIntoRequests,
		for _, readIntoRequest := range readIntoRequests {
			//   6.3.1. Perform readIntoRequest’s close steps, given undefined.
			readIntoRequest.closeSteps(nil)
		}
	}

	// 7. Let sourceCancelPromise be ! stream.[[controller]].[[CancelSteps]](reason).
	sourceCancelPromise := stream.controller.cancelSteps(reason)

	// 8. Return the result of reacting to sourceCancelPromise with a fulfillment step that returns undefined.
	promise, err := promiseThen(stream.vu.Runtime(), sourceCancelPromise,
		// Mimicking Deno's implementation: https://github.com/denoland/deno/blob/main/ext/web/06_streams.js#L405
		func(goja.Value) {},
		func(err goja.Value) {},
	)
	if err != nil {
		common.Throw(stream.vu.Runtime(), err)
	}

	return promise
}

// close implements the specification's [ReadableStreamClose()] abstract operation.
//
// [ReadableStreamClose()]: https://streams.spec.whatwg.org/#readable-stream-close
func (stream *ReadableStream) close() {
	// 1. Assert: stream.[[state]] is "readable".
	if stream.state != ReadableStreamStateReadable {
		common.Throw(stream.vu.Runtime(), newError(AssertionError, "cannot close a stream that is not readable"))
	}

	// 2. Set stream.[[state]] to "closed".
	stream.state = ReadableStreamStateClosed

	// 3. Let reader be stream.[[reader]].
	reader := stream.reader

	// 4. If reader is undefined, return.
	if reader == nil {
		return
	}

	// 5. Resolve reader.[[closedPromise]] with undefined.
	genericReader, ok := reader.(ReadableStreamGenericReader)
	if !ok {
		common.Throw(stream.vu.Runtime(), newError(AssertionError, "reader is not a ReadableStreamGenericReader"))
	}

	_, resolveFunc, _ := genericReader.GetClosed()
	resolveFunc(goja.Undefined())

	// 6. If reader implements ReadableStreamDefaultReader,
	defaultReader, ok := reader.(*ReadableStreamDefaultReader)
	if ok {
		// 6.1. Let readRequests be reader.[[readRequests]].
		readRequests := defaultReader.readRequests

		// 6.2. Set reader.[[readRequests]] to an empty list.
		defaultReader.readRequests = []ReadRequest{}

		// 6.3. For each readRequest of readRequests,
		for _, readRequest := range readRequests {
			// 6.3.1. Perform readRequest’s close steps.
			readRequest.closeSteps()
		}
	}
}

// error implements the specification's [ReadableStreamError] abstract operation.
//
// [ReadableStreamError]: https://streams.spec.whatwg.org/#readable-stream-error
func (stream *ReadableStream) error(e any) {
	// 1. Assert: stream.[[state]] is "readable".
	if stream.state != ReadableStreamStateReadable {
		common.Throw(stream.vu.Runtime(), newError(AssertionError, "cannot error a stream that is not readable"))
	}

	// 2. Set stream.[[state]] to "errored".
	stream.state = ReadableStreamStateErrored

	// 3. Set stream.[[storedError]] to e.
	stream.storedError = e

	// 4. Let reader be stream.[[reader]].
	reader := stream.reader

	// 5. If reader is undefined, return.
	if reader == nil {
		return
	}

	genericReader, ok := reader.(ReadableStreamGenericReader)
	if !ok {
		common.Throw(stream.vu.Runtime(), newError(AssertionError, "reader is not a ReadableStreamGenericReader"))
	}

	// 6. Reject reader.[[closedPromise]] with e.
	promise, _, rejectFunc := genericReader.GetClosed()
	rejectFunc(e)

	// 7. Set reader.[[closedPromise]].[[PromiseIsHandled]] to true.
	// FIXME: See https://github.com/dop251/goja/issues/565
	var (
		err       error
		doNothing = func(goja.Value) {}
	)
	_, err = promiseThen(stream.vu.Runtime(), promise, doNothing, doNothing)
	if err != nil {
		common.Throw(stream.vu.Runtime(), newError(RuntimeError, err.Error()))
	}

	// 8. If reader implements ReadableStreamDefaultReader,
	defaultReader, ok := reader.(*ReadableStreamDefaultReader)
	if ok {
		// 8.1. Perform ! ReadableStreamDefaultReaderErrorReadRequests(reader, e).
		defaultReader.errorReadRequests(e)
		return
	}

	// 9. OTHERWISE, reader is a ReadableStreamBYOBReader
	// 9.1. Assert: reader implements ReadableStreamBYOBReader.
	byobReader, ok := reader.(*ReadableStreamBYOBReader)
	if !ok {
		common.Throw(stream.vu.Runtime(), newError(AssertionError, "reader is not a ReadableStreamBYOBReader"))
	}

	// 9.2. Perform ! ReadableStreamBYOBReaderErrorReadIntoRequests(reader, e).
	byobReader.errorReadIntoRequests(e)
}

// fulfillReadIntoRequest implements the [ReadableStreamFulfillReadIntoRequest()] algorithm.
//
// [ReadableStreamFulfillReadIntoRequest()]: https://streams.spec.whatwg.org/#readable-stream-fulfill-read-into-request
func (stream *ReadableStream) fulfillReadIntoRequest(chunk any, done bool) {
	// 1. Assert: ! ReadableStreamHasBYOBReader(stream) is true.
	if !stream.hasBYOBReader() {
		common.Throw(stream.vu.Runtime(), newError(AssertionError, "stream does not have a BYOB reader"))
	}

	// 2. Let reader be stream.[[reader]].
	reader, ok := stream.reader.(*ReadableStreamBYOBReader)
	if !ok {
		common.Throw(stream.vu.Runtime(), newError(RuntimeError, "reader is not a ReadableStreamBYOBReader"))
	}

	// 3. Assert: reader.[[readIntoRequests]] is not empty.
	if len(reader.readIntoRequests) == 0 {
		common.Throw(stream.vu.Runtime(), newError(AssertionError, "reader.[[readIntoRequests]] is empty"))
	}

	// 4. Let readIntoRequest be reader.[[readIntoRequests]][0].
	readIntoRequest := reader.readIntoRequests[0]

	// 5. Remove readIntoRequest from reader.[[readIntoRequests]].
	reader.readIntoRequests = reader.readIntoRequests[1:]

	if done {
		// 6. If done is true, perform readIntoRequest’s close steps, given undefined.
		readIntoRequest.closeSteps(nil)
	} else {
		// 7. Otherwise, perform readIntoRequest’s chunk steps, given chunk.
		readIntoRequest.chunkSteps(chunk)
	}
}

// fulfillReadRequest implements the [ReadableStreamFulfillReadRequest()] algorithm.
//
// [ReadableStreamFulfillReadRequest()]: https://streams.spec.whatwg.org/#readable-stream-fulfill-read-request
func (stream *ReadableStream) fulfillReadRequest(chunk any, done bool) {
	// 1. Assert: ! ReadableStreamHasDefaultReader(stream) is true.
	if !stream.hasDefaultReader() {
		common.Throw(stream.vu.Runtime(), newError(AssertionError, "stream does not have a default reader"))
	}

	// 2. Let reader be stream.[[reader]].
	reader, ok := stream.reader.(*ReadableStreamDefaultReader)
	if !ok {
		common.Throw(stream.vu.Runtime(), newError(RuntimeError, "reader is not a ReadableStreamDefaultReader"))
	}

	// 3. Assert: reader.[[readRequests]] is not empty.
	if len(reader.readRequests) == 0 {
		common.Throw(stream.vu.Runtime(), newError(AssertionError, "reader.[[readRequests]] is empty"))
	}

	// 4. Let readRequest be reader.[[readRequests]][0].
	readRequest := reader.readRequests[0]

	// 5. Remove readRequest from reader.[[readRequests]].
	reader.readRequests = reader.readRequests[1:]

	if done {
		// 6. If done is true, perform readRequest’s close steps.
		readRequest.closeSteps()
	} else {
		// 7. Otherwise, perform readRequest’s chunk steps, given chunk.
		readRequest.chunkSteps(chunk)
	}
}

// getNumReadIntoRequests implements the [ReadableStreamGetNumReadIntoRequests()] algorithm.
//
// [ReadableStreamGetNumReadIntoRequests()]: https://streams.spec.whatwg.org/#readable-stream-get-num-read-into-requests
func (stream *ReadableStream) getNumReadIntoRequests() int {
	// 1.
	if !stream.hasBYOBReader() {
		common.Throw(stream.vu.Runtime(), newError(AssertionError, "stream does not have a BYOB reader"))
	}

	byobReader, ok := stream.reader.(ReadableStreamBYOBReader)
	if !ok {
		common.Throw(stream.vu.Runtime(), newError(RuntimeError, "reader is not a ReadableStreamBYOBReader"))
	}

	// 2.
	return len(byobReader.readIntoRequests)
}

// getNumReadRequests implements the [ReadableStreamGetNumReadRequests()] algorithm.
//
// [ReadableStreamGetNumReadRequests()]: https://streams.spec.whatwg.org/#readable-stream-get-num-read-requests
func (stream *ReadableStream) getNumReadRequests() int {
	// 1. Assert: ! ReadableStreamHasDefaultReader(stream) is true.
	if !stream.hasDefaultReader() {
		common.Throw(stream.vu.Runtime(), newError(AssertionError, "stream does not have a default reader"))
	}

	// 2. Return stream.[[reader]].[[readRequests]]'s size.
	defaultReader, ok := stream.reader.(*ReadableStreamDefaultReader)
	if !ok {
		common.Throw(stream.vu.Runtime(), newError(RuntimeError, "reader is not a ReadableStreamDefaultReader"))
	}

	return len(defaultReader.readRequests)
}

// hasBYOBReader implements the [ReadableStreamHasBYOBReader()] algorithm.
//
// [ReadableStreamHasBYOBReader()]: https://streams.spec.whatwg.org/#readable-stream-has-byob-reader
func (stream *ReadableStream) hasBYOBReader() bool {
	// 1. Let reader be stream.[[reader]].
	reader := stream.reader

	// 2. If reader is undefined, return false.
	if reader == nil {
		return false
	}

	// 3. If reader implements ReadableStreamBYOBReader, return true.
	_, ok := reader.(ReadableStreamBYOBReader)
	return ok
}

// hasDefaultReader implements the [ReadableStreamHasDefaultReader()] algorithm.
//
// [ReadableStreamHasDefaultReader()]: https://streams.spec.whatwg.org/#readable-stream-has-default-reader
func (stream *ReadableStream) hasDefaultReader() bool {
	// 1. Let reader be stream.[[reader]].
	reader := stream.reader

	// 2. If reader is undefined, return false.
	if reader == nil {
		return false
	}

	// 3. If reader implements ReadableStreamDefaultReader, return true.
	_, ok := reader.(*ReadableStreamDefaultReader)
	return ok
}
