// BSD 2-Clause License
//
// Copyright (c) 2020 Don Owens <don@regexguy.com>.  All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice,
//   this list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

// The fileutil package provides utilities for working with files and io
// streams in Go.
//
// Installation
//
//   go get github.com/cuberat-go/fileutil
package fileutil

import (
    // Built-in/core modules.
    "bufio"
    "bytes"
    bzip2 "compress/bzip2"
    "errors"
    exec "os/exec"
    "fmt"
    gzip "compress/gzip"
    "io"
    "os"
    "strings"

    // Third-party modules.


    // First-party modules.
)

var (
    Err_UnknownSuffix error = errors.New("Unknown suffix")
)

type CloseFunc func() error

// An interface for an `io.WriteCloser` that also has a name, similar to a
// *os.FILE.
type NameWriteCloser interface {
    Name() string
    Write(p []byte) (int, error)
    Close() error
}

// An interface for an `io.ReadCloser` that also has a name, similar to a
// *os.FILE.
type NameReadCloser interface {
    io.ReadCloser
    Name() string
}

type read_closer struct {
    r io.Reader
    close_func CloseFunc
}

func (w *read_closer) Read(p []byte) (n int, err error) {
    return w.r.Read(p)
}

func (w *read_closer) Close() error {
    if w.close_func == nil {
        return nil
    }
    return w.close_func()
}

// Given an `io.Reader`, return an `io.ReadCloser` with the provided `Close()`
// function.
func ReadCloserFromReader(r io.Reader, close_func CloseFunc) io.ReadCloser {
    return &read_closer{r: r, close_func: close_func}
}

type name_read_closer struct {
    name string
    rc io.ReadCloser
}

func (w *name_read_closer) Name() string {
    return w.name
}

func (w *name_read_closer) Read(p []byte) (n int, err error) {
    return w.rc.Read(p)
}

func (w *name_read_closer) Close() error {
    return w.rc.Close()
}

// Given an `io.ReadCloser`, return a `NameReadCloser` with the provided name.
func NameReadCloserFromReadCloser(
    name string,
    rc io.ReadCloser,
) NameReadCloser {
    return &name_read_closer{name: name, rc: rc}
}

// Given an `io.Reader`, return a `NameReadCloser` with the provided name and
// `Close()` function.
func NameReadCloserFromReader(
    name string,
    r io.Reader,
    close_func CloseFunc,
) NameReadCloser {
    return &name_read_closer{
        name: name,
        rc: ReadCloserFromReader(r, close_func),
    }
}

// Given an `io.WriteCloser`, return a `NameWriteCloser` with the provided
// name.
func NameWriteCloserFromWriteCloser(
    name string,
    wc io.WriteCloser,
) NameWriteCloser {
    return &name_write_closer{name: name, writer: wc, close_func: wc.Close}
}

type name_write_closer struct {
    close_func CloseFunc
    name string
    writer io.Writer
}

// Returns a `NameWriteCloser` with the provided name and `Close()` function
// for the provided `io.Writer`.
func NameWriteCloserFromWriter(
    name string,
    writer io.Writer,
    close_func CloseFunc,
) NameWriteCloser {
    return &name_write_closer{
        name: name,
        writer: writer,
        close_func: close_func,
    }
}

func (w *name_write_closer) Close() error {
    if w.close_func == nil {
        return nil
    }
    return w.close_func()
}

func (w *name_write_closer) Name() string {
    return w.name
}

func (w *name_write_closer) Write(p []byte) (int, error) {
    return w.writer.Write(p)
}

type write_closer struct {
    writer io.Writer
    close_func CloseFunc
}

func (w *write_closer) Close() error {
    if w.close_func == nil {
        return nil
    }
    return w.close_func()
}

func (w *write_closer) Write(p []byte) (int, error) {
    return w.writer.Write(p)
}

// Given an `io.Writer`, returns an `io.WriteCloser` that calls the specified
// `Close()` function when `Close()` is called on the `io.WriteCloser`.
func WriteCloserFromWriter(
    writer io.Writer,
    close_func CloseFunc,
) io.WriteCloser {
    return &write_closer{writer: writer, close_func: close_func}
}

// Shortcut for calling `CreateFileBuffered()` with the default buffer size.
// Equivalent to `CreateFileBuffered()` with 0 as the size parameter.
func CreateFile(outfile string) (NameWriteCloser, error) {
    return CreateFileBuffered(outfile, 0)
}

// Non-buffered version of `CreateFile()` and `CreateFileBuffered()`.
// Equivalent to `CreateFileBuffered()` with -1 as the size parameter.
func CreateFileSync(outfile string) (NameWriteCloser, error) {
    return CreateFileBuffered(outfile, -1)
}

// Opens a file for writing (buffered). The `size` argument indicates that the
// underlying buffer should be at least `size` bytes. If `size` < 0, open the
// file with no buffering. If `size` == 0, a size of 16K is used. If the file
// name ends in a supported compression suffix, output will be compressed in
// that format.
//
// Supported compression:
//    gzip  (.gz)
//    bzip2 (.bz2) -- calls external program
//    xz    (.xz)  -- calls external program
//
// Be sure to call `Close()` explicitly to flush any buffers and properly shut
// down any compression layers.
func CreateFileBuffered(outfile string, size int) (NameWriteCloser, error) {
    if size == 0 {
        size = 16384
    }

    out_fh, err := os.Create(outfile)
    if err != nil {
        return nil, fmt.Errorf("couldn't open output file %s: %w",
            outfile, err)
    }

    idx := strings.LastIndex(outfile, ".")
    if idx <= -1 || idx >= len(outfile) - 1 {
        // No file extension, so no compression layer required.
        if size > 0 {
            return NameWriteCloserFromWriteCloser(outfile,
                add_buffer(out_fh, size)), nil
        }
        return out_fh, nil
    }

    suffix := outfile[idx+1:len(outfile)]

    w, err := AddCompressionLayer(out_fh, suffix)
    if err != nil {
        if err == Err_UnknownSuffix {
            // No compression layer added
            if size > 0 {
                return NameWriteCloserFromWriteCloser(outfile,
                    add_buffer(out_fh, size)), nil
            }
            return out_fh, nil
        } else {
            out_fh.Close()
            return nil, fmt.Errorf("couldn't add compression layer: %w", err)
        }
    }

    if size > 0 {
        w = add_buffer(w, size)
    }

    return NameWriteCloserFromWriteCloser(outfile, w), nil
}

func add_buffer(w_orig io.WriteCloser, size int) io.WriteCloser {
    w_buffered := bufio.NewWriterSize(w_orig, size)

    close_func := func() error {
        var close_err error
        if close_err = w_buffered.Flush(); close_err != nil {
            return close_err
        }

        if close_err = w_orig.Close(); close_err != nil {
            return close_err
        }

        return nil
    }

    return WriteCloserFromWriter(w_buffered, close_func)
}

// Opens a file in read-only mode. If the file name ends in a supported
// compression suffix, input will be decompressed.
//
// Supported decompression:
//    gzip  (.gz)
//    bzip2 (.bz2)
//    xz    (.xz) -- calls external program
//
// Call `Close()` on the returned NameReadCloser to avoid leaking filehandles
// and to properly shut down any compression layers.
func OpenFile(infile string) (NameReadCloser, error) {
    in_fh, err := os.Open(infile)
    if err != nil {
        return nil, err
    }

    idx := strings.LastIndex(infile, ".")
    if idx <= -1 || idx >= len(infile) - 1 {
        return in_fh, nil
    }

    suffix := infile[idx+1:len(infile)]

    r, err := AddDecompressionLayer(in_fh, suffix)
    if err != nil {
        if err == Err_UnknownSuffix {
            return in_fh, nil
        } else {
            in_fh.Close()
            return nil, fmt.Errorf("couldn't add decompression layer: %w",
                err)
        }
    }

    close_func := func() error {
        r.Close()
        return in_fh.Close()
    }

    return NameReadCloserFromReadCloser(infile,
        ReadCloserFromReader(r, close_func)), nil
}

// Adds decompression to input read from reader r, if the suffix is supported.
//
// Supported decompression:
//    gzip  (gz)
//    bzip2 (bz2)
//    xz    (xz) -- calls external program
func AddDecompressionLayer(
    r io.Reader,
    suffix string,
) (io.ReadCloser, error) {
    switch suffix {
    case "gz", "gzip":
        new_reader, err := gzip.NewReader(r)
        if err != nil {
            return nil, fmt.Errorf("couldn't create gzip reader: %w", err)
        }

        close_func := func() error {
            return new_reader.Close()
        }

        return ReadCloserFromReader(new_reader, close_func), nil

    case "bz2", "bzip2":
        new_reader := bzip2.NewReader(r)
        close_func := func() error { return nil }
        return ReadCloserFromReader(new_reader, close_func), nil

    case "xz":
        return new_xz_reader(r)
    }

    return nil, Err_UnknownSuffix
}

// Adds compression to output written to writer w, if the suffix is supported.
//
// Supported compression:
//    gzip  (gz)
//    bzip2 (bz2) -- calls external program
//    xz    (xz)  -- calls external program
//
// Call the Close() method on the returned io.WriteCloser to properly shutdown
// the compression layer.
func AddCompressionLayer(
    w io.WriteCloser,
    suffix string,
) (
    io.WriteCloser,
    error,
) {

    switch suffix {
    case "gz", "gzip":
        gzip_writer, err := gzip.NewWriterLevel(w, gzip.BestCompression)
        if err != nil {
            return nil, fmt.Errorf("couldn't create gzip writer: %w", err)
        }

        close_func := func() error {
            gzip_writer.Flush()
            return gzip_writer.Close()
        }

        return WriteCloserFromWriter(gzip_writer, close_func), nil

    case "bz2", "bzip2":
        return new_bz2_writer(w)

    case "xz":
        return new_xz_writer(w)
    }

    return nil, Err_UnknownSuffix
}

func get_writer_pipe_from_exec_with_writer(prog_stdout io.Writer,
    prog ...string) (io.WriteCloser, error) {

    name := prog[0]
    args := prog[1:]
    cmd := exec.Command(name, args...)
    cmd.Stdout = prog_stdout

    writer_closer, err := cmd.StdinPipe()
    if err != nil {
        return nil,
        fmt.Errorf("couldn't get stdout pipe in prog writer (%s): %w",
            strings.Join(prog, " "), err)
    }

    err = cmd.Start()
    if err != nil {
        writer_closer.Close()
        return nil, fmt.Errorf("couldn't start process %s: %w",
            strings.Join(prog, " "), err)
    }

    close_func := func() error {
        writer_closer.Close()
        return format_exit_error(cmd.Wait())
    }

    return WriteCloserFromWriter(writer_closer, close_func), nil
}

func format_exit_error(orig_err error) error {
    if exit_err, ok := orig_err.(*exec.ExitError); ok {
        errs := make([]string, 0, 2)
        proc_state := exit_err.ProcessState
        if proc_state != nil {
            errs = append(errs, fmt.Sprintf("process exited with code %d: ",
                proc_state.ExitCode()))
        }
        if len(exit_err.Stderr) > 0 {
            buf := bytes.NewBuffer(exit_err.Stderr)
            first_line, err := buf.ReadString('\n')
            if len(first_line) == 0 && err != nil {
                first_line, _ = buf.ReadString('\n')
            }
            errs = append(errs, first_line)
        }
        if len(errs) == 0 {
            return orig_err
        }
        return fmt.Errorf("%s", strings.Join(errs, ": "))
    }

    return orig_err
}

func get_reader_pipe_from_exec_with_reader(prog_stdin io.Reader,
    prog ...string) (io.ReadCloser, error) {

    name := prog[0]
    args := prog[1:]
    cmd := exec.Command(name, args...)
    cmd.Stdin = prog_stdin
    reader_closer, err := cmd.StdoutPipe()
    if err != nil {
        return nil, fmt.Errorf("couldn't get stdout pipe in prog reader (%s): %w",
            strings.Join(prog, " "), err)
    }

    err = cmd.Start()
    if err != nil {
        reader_closer.Close()
        return nil, fmt.Errorf("couldn't start process %s: %w",
            strings.Join(prog, " "), err)
    }

    close_func := func() error {
        reader_closer.Close()
        return format_exit_error(cmd.Wait())
    }

    return ReadCloserFromReader(reader_closer, close_func), nil
}

func new_bz2_writer(w io.Writer) (io.WriteCloser, error) {
    path, err := find_exec("bzip2")
    if err !=  nil {
        return nil, err
    }

    return get_writer_pipe_from_exec_with_writer(w, path, "-z", "-c")
}

func new_xz_writer(w io.Writer) (io.WriteCloser, error) {
    xz_path, err := find_exec("xz")
    if err !=  nil {
        return nil, err
    }

    return get_writer_pipe_from_exec_with_writer(w, xz_path, "-z", "-e", "-c")
}

func new_xz_reader(r io.Reader) (io.ReadCloser, error) {
    xz_path, err := find_exec("xz")
    if err !=  nil {
        return nil, err
    }

    return get_reader_pipe_from_exec_with_reader(r, xz_path, "-d", "-c")
}

func find_exec(file string) (string, error) {
    dirs := []string{"/bin", "/usr/bin", "/usr/local/bin"}

    for _, dir := range dirs {
        path := fmt.Sprintf("%s/%s", dir, file)
        _, err := os.Stat(path)
        if err == nil {
            return path, nil
        }
    }

    return "", fmt.Errorf("couldn't find executable %s", file)
}

// Runs the list of commands, piping the output of each one to the next. The
// output of the last command is sent to the final_writer passed in.
// Each command is represented as a slice of strings. The first element of the
// slice should be the full path to the program to run. The remaining elements
// of the slice should be the arguments to the program.
//
// The writer returned writes to the standard input of the first program in
// the list. Close() should be called when writing has been completed.
func OpenPipesToWriter(final_writer io.Writer,
    progs [][]string) (io.WriteCloser, error) {

    overall_close_func := func() error { return nil }
    writer := final_writer

    last := len(progs) - 1
    for i := range progs {
        close_func := overall_close_func
        prog := progs[last - i]
        new_write_closer, err :=
            get_writer_pipe_from_exec_with_writer(writer, prog...)
        if err != nil {
            overall_close_func()
            return nil, err
        }

        overall_close_func = func() error {
            err1 := new_write_closer.Close()
            err2 := close_func()
            if err1 != nil {
                return err1
            }
            return err2
        }

        writer = new_write_closer
    }

    return WriteCloserFromWriter(writer, overall_close_func), nil
}
