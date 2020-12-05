

# fileutil
`import "github.com/cuberat-go/fileutil"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>
The fileutil package provides utilities for working with files and io
streams in Go.

Installation


	go get github.com/cuberat-go/fileutil




## <a name="pkg-index">Index</a>
* [Variables](#pkg-variables)
* [func AddCompressionLayer(w io.WriteCloser, suffix string) (io.WriteCloser, error)](#AddCompressionLayer)
* [func AddDecompressionLayer(r io.Reader, suffix string) (io.ReadCloser, error)](#AddDecompressionLayer)
* [func OpenPipesToWriter(final_writer io.Writer, progs [][]string) (io.WriteCloser, error)](#OpenPipesToWriter)
* [func ReadCloserFromReader(r io.Reader, close_func CloseFunc) io.ReadCloser](#ReadCloserFromReader)
* [func WriteCloserFromWriter(writer io.Writer, close_func CloseFunc) io.WriteCloser](#WriteCloserFromWriter)
* [type CloseFunc](#CloseFunc)
* [type NameReadCloser](#NameReadCloser)
  * [func NameReadCloserFromReadCloser(name string, rc io.ReadCloser) NameReadCloser](#NameReadCloserFromReadCloser)
  * [func NameReadCloserFromReader(name string, r io.Reader, close_func CloseFunc) NameReadCloser](#NameReadCloserFromReader)
  * [func OpenFile(infile string) (NameReadCloser, error)](#OpenFile)
* [type NameWriteCloser](#NameWriteCloser)
  * [func CreateFile(outfile string) (NameWriteCloser, error)](#CreateFile)
  * [func CreateFileBuffered(outfile string, size int) (NameWriteCloser, error)](#CreateFileBuffered)
  * [func CreateFileSync(outfile string) (NameWriteCloser, error)](#CreateFileSync)
  * [func NameWriteCloserFromWriteCloser(name string, wc io.WriteCloser) NameWriteCloser](#NameWriteCloserFromWriteCloser)
  * [func NameWriteCloserFromWriter(name string, writer io.Writer, close_func CloseFunc) NameWriteCloser](#NameWriteCloserFromWriter)


#### <a name="pkg-files">Package files</a>
[fileutil.go](/src/github.com/cuberat-go/fileutil/fileutil.go) 



## <a name="pkg-variables">Variables</a>
``` go
var (
    Err_UnknownSuffix error = errors.New("Unknown suffix")
)
```


## <a name="AddCompressionLayer">func</a> [AddCompressionLayer](/src/target/fileutil.go?s=10362:10465#L370)
``` go
func AddCompressionLayer(
    w io.WriteCloser,
    suffix string,
) (
    io.WriteCloser,
    error,
)
```
Adds compression to output written to writer w, if the suffix is supported.

Supported compression:


	gzip  (gz)
	bzip2 (bz2) -- calls external program
	xz    (xz)  -- calls external program

Call the Close() method on the returned io.WriteCloser to properly shutdown
the compression layer.



## <a name="AddDecompressionLayer">func</a> [AddDecompressionLayer](/src/target/fileutil.go?s=9327:9415#L332)
``` go
func AddDecompressionLayer(
    r io.Reader,
    suffix string,
) (io.ReadCloser, error)
```
Adds decompression to input read from reader r, if the suffix is supported.

Supported decompression:


	gzip  (gz)
	bzip2 (bz2)
	xz    (xz) -- calls external program



## <a name="OpenPipesToWriter">func</a> [OpenPipesToWriter](/src/target/fileutil.go?s=14948:15040#L534)
``` go
func OpenPipesToWriter(final_writer io.Writer,
    progs [][]string) (io.WriteCloser, error)
```
Runs the list of commands, piping the output of each one to the next. The
output of the last command is sent to the final_writer passed in.
Each command is represented as a slice of strings. The first element of the
slice should be the full path to the program to run. The remaining elements
of the slice should be the arguments to the program.

The writer returned writes to the standard input of the first program in
the list. Close() should be called when writing has been completed.



## <a name="ReadCloserFromReader">func</a> [ReadCloserFromReader](/src/target/fileutil.go?s=2645:2719#L83)
``` go
func ReadCloserFromReader(r io.Reader, close_func CloseFunc) io.ReadCloser
```
Given an `io.Reader`, return an `io.ReadCloser` with the provided `Close()`
function.



## <a name="WriteCloserFromWriter">func</a> [WriteCloserFromWriter](/src/target/fileutil.go?s=5034:5126#L187)
``` go
func WriteCloserFromWriter(
    writer io.Writer,
    close_func CloseFunc,
) io.WriteCloser
```
Given an `io.Writer`, returns an `io.WriteCloser` that calls the specified
`Close()` function when `Close()` is called on the `io.WriteCloser`.




## <a name="CloseFunc">type</a> [CloseFunc](/src/target/fileutil.go?s=1898:1925#L48)
``` go
type CloseFunc func() error
```









## <a name="NameReadCloser">type</a> [NameReadCloser](/src/target/fileutil.go?s=2207:2276#L60)
``` go
type NameReadCloser interface {
    io.ReadCloser
    Name() string
}
```
An interface for an `io.ReadCloser` that also has a name, similar to a
*os.FILE.







### <a name="NameReadCloserFromReadCloser">func</a> [NameReadCloserFromReadCloser](/src/target/fileutil.go?s=3153:3243#L105)
``` go
func NameReadCloserFromReadCloser(
    name string,
    rc io.ReadCloser,
) NameReadCloser
```
Given an `io.ReadCloser`, return a `NameReadCloser` with the provided name.


### <a name="NameReadCloserFromReader">func</a> [NameReadCloserFromReader](/src/target/fileutil.go?s=3399:3506#L114)
``` go
func NameReadCloserFromReader(
    name string,
    r io.Reader,
    close_func CloseFunc,
) NameReadCloser
```
Given an `io.Reader`, return a `NameReadCloser` with the provided name and
`Close()` function.


### <a name="OpenFile">func</a> [OpenFile](/src/target/fileutil.go?s=8355:8407#L293)
``` go
func OpenFile(infile string) (NameReadCloser, error)
```
Opens a file in read-only mode. If the file name ends in a supported
compression suffix, input will be decompressed.

Supported decompression:


	gzip  (.gz)
	bzip2 (.bz2)
	xz    (.xz) -- calls external program

Call `Close()` on the returned NameReadCloser to avoid leaking filehandles
and to properly shut down any compression layers.





## <a name="NameWriteCloser">type</a> [NameWriteCloser](/src/target/fileutil.go?s=2015:2118#L52)
``` go
type NameWriteCloser interface {
    Name() string
    Write(p []byte) (int, error)
    Close() error
}
```
An interface for an `io.WriteCloser` that also has a name, similar to a
*os.FILE.







### <a name="CreateFile">func</a> [CreateFile](/src/target/fileutil.go?s=5344:5400#L196)
``` go
func CreateFile(outfile string) (NameWriteCloser, error)
```
Shortcut for calling `CreateFileBuffered()` with the default buffer size.
Equivalent to `CreateFileBuffered()` with 0 as the size parameter.


### <a name="CreateFileBuffered">func</a> [CreateFileBuffered](/src/target/fileutil.go?s=6277:6351#L219)
``` go
func CreateFileBuffered(outfile string, size int) (NameWriteCloser, error)
```
Opens a file for writing (buffered). The `size` argument indicates that the
underlying buffer should be at least `size` bytes. If `size` < 0, open the
file with no buffering. If `size` == 0, a size of 16K is used. If the file
name ends in a supported compression suffix, output will be compressed in
that format.

Supported compression:


	gzip  (.gz)
	bzip2 (.bz2) -- calls external program
	xz    (.xz)  -- calls external program

Be sure to call `Close()` explicitly to flush any buffers and properly shut
down any compression layers.


### <a name="CreateFileSync">func</a> [CreateFileSync](/src/target/fileutil.go?s=5589:5649#L202)
``` go
func CreateFileSync(outfile string) (NameWriteCloser, error)
```
Non-buffered version of `CreateFile()` and `CreateFileBuffered()`.
Equivalent to `CreateFileBuffered()` with -1 as the size parameter.


### <a name="NameWriteCloserFromWriteCloser">func</a> [NameWriteCloserFromWriteCloser](/src/target/fileutil.go?s=3701:3795#L127)
``` go
func NameWriteCloserFromWriteCloser(
    name string,
    wc io.WriteCloser,
) NameWriteCloser
```
Given an `io.WriteCloser`, return a `NameWriteCloser` with the provided
name.


### <a name="NameWriteCloserFromWriter">func</a> [NameWriteCloserFromWriter](/src/target/fileutil.go?s=4084:4198#L142)
``` go
func NameWriteCloserFromWriter(
    name string,
    writer io.Writer,
    close_func CloseFunc,
) NameWriteCloser
```
Returns a `NameWriteCloser` with the provided name and `Close()` function
for the provided `io.Writer`.









- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
