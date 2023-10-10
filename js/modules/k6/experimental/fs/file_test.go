package fs

import (
	"bytes"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileImpl(t *testing.T) {
	t.Parallel()

	t.Run("stat on closed file should fail", func(t *testing.T) {
		t.Parallel()

		f := &file{
			path:   "/bonjour.txt",
			data:   []byte("hello"),
			offset: 0,
			closed: atomic.Bool{},
		}

		err := f.Close()
		require.NoError(t, err)

		gotInfo, gotErr := f.stat()

		assert.Nil(t, gotInfo)
		assert.Error(t, gotErr)

		var fsErr *fsError
		assert.ErrorAs(t, gotErr, &fsErr)
		assert.Equal(t, BadResourceError, fsErr.kind)
	})

	t.Run("read", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name     string
			into     []byte
			fileData []byte
			offset   int
			wantInto []byte
			wantN    int
			wantErr  errorKind
		}{
			{
				name:     "reading the entire file into a buffer fitting the whole file should succeed",
				into:     make([]byte, 5),
				fileData: []byte("hello"),
				offset:   0,
				wantInto: []byte("hello"),
				wantN:    5,
				wantErr:  0, // No error expected
			},
			{
				name:     "reading a file larger than the provided buffer should succeed",
				into:     make([]byte, 3),
				fileData: []byte("hello"),
				offset:   0,
				wantInto: []byte("hel"),
				wantN:    3,
				wantErr:  0, // No error expected
			},
			{
				name:     "reading a file larger than the provided buffer at an offset should succeed",
				into:     make([]byte, 3),
				fileData: []byte("hello"),
				offset:   2,
				wantInto: []byte("llo"),
				wantN:    3,
				wantErr:  0, // No error expected
			},
			{
				name:     "reading file data into a zero sized buffer should succeed",
				into:     []byte{},
				fileData: []byte("hello"),
				offset:   0,
				wantInto: []byte{},
				wantN:    0,
				wantErr:  0, // No error expected
			},
			{
				name:     "reading past the end of the file should fill the buffer and fail with EOF",
				into:     make([]byte, 10),
				fileData: []byte("hello"),
				offset:   0,
				wantInto: []byte{'h', 'e', 'l', 'l', 'o', 0, 0, 0, 0, 0},
				wantN:    5,
				wantErr:  EOFError,
			},
			{
				name:     "reading into a prefilled buffer overrides its content",
				into:     []byte("world!"),
				fileData: []byte("hello"),
				offset:   0,
				wantInto: []byte("hello!"),
				wantN:    5,
				wantErr:  EOFError,
			},
			{
				name:     "reading an empty file should fail with EOF",
				into:     make([]byte, 10),
				fileData: []byte{},
				offset:   0,
				wantInto: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				wantN:    0,
				wantErr:  EOFError,
			},
			{
				name: "reading from the end of a file should fail with EOF",
				into: make([]byte, 10),
				// Note that the offset is larger than the file size
				fileData: []byte("hello"),
				offset:   5,
				wantInto: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				wantN:    0,
				wantErr:  EOFError,
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				f := &file{
					path:   "",
					data:   tc.fileData,
					offset: tc.offset,
				}

				gotN, err := f.Read(tc.into)

				// Cast the error to your custom error type to access its kind
				var gotErr errorKind
				if err != nil {
					var fsErr *fsError
					ok := errors.As(err, &fsErr)
					if !ok {
						t.Fatalf("unexpected error type: got %T, want %T", err, &fsError{})
					}
					gotErr = fsErr.kind
				}

				if gotN != tc.wantN || gotErr != tc.wantErr {
					t.Errorf("Read() = %d, %v, want %d, %v", gotN, gotErr, tc.wantN, tc.wantErr)
				}

				if !bytes.Equal(tc.into, tc.wantInto) {
					t.Errorf("Read() into = %v, want %v", tc.into, tc.wantInto)
				}
			})
		}
	})

	t.Run("read from closed file should fail", func(t *testing.T) {
		t.Parallel()

		f := &file{
			path:   "/bonjour.txt",
			data:   []byte("hello"),
			offset: 0,
			closed: atomic.Bool{},
		}

		err := f.Close()
		require.NoError(t, err)

		gotN, gotErr := f.Read(make([]byte, 10))

		assert.Equal(t, 0, gotN)
		assert.Error(t, gotErr)

		var fsErr *fsError
		assert.ErrorAs(t, gotErr, &fsErr)
		assert.Equal(t, BadResourceError, fsErr.kind)
	})

	t.Run("seek", func(t *testing.T) {
		t.Parallel()

		type args struct {
			offset int
			whence SeekMode
		}

		// The test file is 100 bytes long
		tests := []struct {
			name       string
			fileOffset int
			args       args
			wantOffset int
			wantError  bool
		}{
			{
				name:       "seek using SeekModeStart within file bounds should succeed",
				fileOffset: 0,
				args:       args{50, SeekModeStart},
				wantOffset: 50,
				wantError:  false,
			},
			{
				name:       "seek using SeekModeStart beyond file boundaries should fail",
				fileOffset: 0,
				args:       args{150, SeekModeStart},
				wantOffset: 0,
				wantError:  true,
			},
			{
				name:       "seek using SeekModeStart and a negative offset should fail",
				fileOffset: 0,
				args:       args{-50, SeekModeStart},
				wantOffset: 0,
				wantError:  true,
			},
			{
				name:       "seek using SeekModeCurrent within file bounds at offset 0 should succeed",
				fileOffset: 0,
				args:       args{10, SeekModeCurrent},
				wantOffset: 10,
				wantError:  false,
			},
			{
				name:       "seek using SeekModeCurrent within file bounds at non-zero offset should succeed",
				fileOffset: 20,
				args:       args{10, SeekModeCurrent},
				wantOffset: 30,
				wantError:  false,
			},
			{
				name:       "seek using SeekModeCurrent beyond file boundaries should fail",
				fileOffset: 20,
				args:       args{100, SeekModeCurrent},
				wantOffset: 20,
				wantError:  true,
			},
			{
				name:       "seek using SeekModeCurrent and a negative offset should succeed",
				fileOffset: 20,
				args:       args{-10, SeekModeCurrent},
				wantOffset: 10,
				wantError:  false,
			},
			{
				name:       "seek using SeekModeCurrent and an out of bound negative offset should fail",
				fileOffset: 20,
				args:       args{-40, SeekModeCurrent},
				wantOffset: 20,
				wantError:  true,
			},
			{
				name:       "seek using SeekModeEnd within file bounds should succeed",
				fileOffset: 20,
				args:       args{-20, SeekModeEnd},
				wantOffset: 80,
				wantError:  false,
			},
			{
				name:       "seek using SeekModeEnd beyond file start should fail",
				fileOffset: 20,
				args:       args{-110, SeekModeEnd},
				wantOffset: 20,
				wantError:  true, // File is 100 bytes long
			},
			{
				name:       "seek using SeekModeEnd and a positive offset should fail",
				fileOffset: 20,
				args:       args{10, SeekModeEnd},
				wantOffset: 20,
				wantError:  true,
			},
			{
				name:       "seek with invalid whence should fail",
				fileOffset: 0,
				args:       args{10, SeekMode(42)},
				wantOffset: 0,
				wantError:  true,
			},
		}

		for _, tt := range tests {
			tt := tt

			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				f := &file{data: make([]byte, 100), offset: tt.fileOffset}

				got, err := f.Seek(tt.args.offset, tt.args.whence)
				if tt.wantError {
					assert.Error(t, err, tt.name)
				} else {
					assert.NoError(t, err, tt.name)
					assert.Equal(t, tt.wantOffset, got, tt.name)
				}
			})
		}
	})

	t.Run("seek on closed file should fail", func(t *testing.T) {
		t.Parallel()

		f := &file{
			path:   "/bonjour.txt",
			data:   []byte("hello"),
			offset: 0,
			closed: atomic.Bool{},
		}

		err := f.Close()
		require.NoError(t, err)

		gotN, gotErr := f.Seek(0, SeekModeStart)

		assert.Equal(t, 0, gotN)
		assert.Error(t, gotErr)

		var fsErr *fsError
		assert.ErrorAs(t, gotErr, &fsErr)
		assert.Equal(t, BadResourceError, fsErr.kind)
	})

	t.Run("close", func(t *testing.T) {
		t.Parallel()

		t.Run("closing a file should succeed", func(t *testing.T) {
			t.Parallel()

			f := &file{
				path:   "/bonjour.txt",
				data:   []byte("hello"),
				offset: 0,
				// defaults to closed == false
			}

			gotErr := f.Close()

			assert.NoError(t, gotErr)
			assert.Equal(t, true, f.closed.Load())
		})

		t.Run("double closing a file should fail", func(t *testing.T) {
			t.Parallel()

			f := &file{
				path:   "/bonjour.txt",
				data:   []byte("hello"),
				offset: 0,
				closed: atomic.Bool{},
			}

			err := f.Close()
			require.NoError(t, err)
			gotErr := f.Close()

			var fsErr *fsError
			assert.ErrorAs(t, gotErr, &fsErr)
			assert.Equal(t, BadResourceError, fsErr.kind)
		})
	})
}
