package fileutil_test

import (
    // Built-in/core modules.
    "fmt"
    ioutil "io/ioutil"
    "os"
    "path"
    "strings"
    "testing"

    // Third-party modules.


    // First-party modules.
    fileutil "github.com/cuberat-go/fileutil"
)

func TestFileCreation(t *testing.T) {
    tests := []struct {
        Name string
        Suffix string
        Magic string
        BufSize int
    }{
        // default buffer size.
        {"bzip2", ".bz2", "BZh", 0},
        {"gzip", ".gz", "\x1F\x8B", 0},
        {"xz", ".xz", "\xFD\x37\x7A\x58\x5A\x00", 0},
        {"plain", ".txt", "", 0},

        // no buffer (synchronous)
        {"bzip2", ".bz2", "BZh", -1},
        {"gzip", ".gz", "\x1F\x8B", -1},
        {"xz", ".xz", "\xFD\x37\x7A\x58\x5A\x00", -1},
        {"plain", ".txt", "", -1},

        // custom buffer size
        {"bzip2", ".bz2", "BZh", 32},
        {"gzip", ".gz", "\x1F\x8B", 32},
        {"xz", ".xz", "\xFD\x37\x7A\x58\x5A\x00", 32},
        {"plain", ".txt", "", 32},
    }

    for _, test := range tests {
        name := fmt.Sprintf("%s_buf_%d", test.Name, test.BufSize)
        t.Run(name, func(st *testing.T) {
            run_file_create_test(st, test.Name, test.Suffix, test.Magic,
                test.BufSize)
        })
    }
}

func run_file_create_test(
    t *testing.T,
    compress_name,
    suffix,
    magic string,
    buffer_size int,
) {
    out_dir, err := ioutil.TempDir("", "fileutil_test_*")
    if err != nil {
        t.Errorf("couldn't create temp directory for testing: %s", err)
        return
    }
    defer os.RemoveAll(out_dir)

    test_str := `orem ipsum dolor sit amet, consectetur adipiscing elit. Aliquam mattis elementum lobortis. Integer ut nibh in odio sagittis viverra in rhoncus purus. Curabitur aliquam finibus massa, porttitor pretium massa rutrum et. Nam sodales dui viverra odio fermentum, in dictum ex tincidunt. Nunc vel tempor erat. Vivamus pharetra vestibulum felis et.
`

    t.Logf("created temp dir %q", out_dir)

    file := path.Join(out_dir, "test_out" + suffix)
    out_fh, err := fileutil.CreateFileBuffered(file, buffer_size)
    if err != nil {
        if strings.Contains(err.Error(), "couldn't find executable") {
            t.Skip(
                fmt.Sprintf("%s compression not suported on this system: %s",
                    compress_name, err.Error()),
            )
            return
        }
        t.Errorf("couldn't open output file %q: %s", file, err)
        return
    }

    fmt.Fprintf(out_fh, "%s", test_str)
    out_fh.Close()

    t.Logf("wrote to %q", file)

    data_bytes := make([]byte, len(magic))

    if len(magic) > 0 {
        in_fh, err := os.Open(file)
        _, err = in_fh.Read(data_bytes)
        in_fh.Close()
        if err != nil {
            t.Errorf("couldn't read %s magic number from %q: %s", compress_name,
                file, err)
            return
        }

        if string(data_bytes) != magic {
            t.Errorf("magic number %q incorrect for %s. Expected %q",
                data_bytes, compress_name, magic)
            return
        }
    }

    in, err := fileutil.OpenFile(file)
    if err != nil {
        t.Errorf("couldn't open file %q for input: %s", file, err)
        return
    }
    defer in.Close()

    data_bytes, err = ioutil.ReadAll(in)
    if err != nil {
        t.Errorf("couldn't read all from file %q: %s", file, err)
        return
    }

    if string(data_bytes) != test_str {
        t.Errorf("file contents incorrect: got %q, expected %q",
            string(data_bytes), test_str)
        return
    }
}

func TestPipesWriter(t *testing.T) {
    out_dir, err := ioutil.TempDir("", "fileutil_test_*")
    if err != nil {
        t.Errorf("couldn't create temp directory for testing: %s", err)
        return
    }
    t.Logf("created temp dir %q", out_dir)
    defer os.RemoveAll(out_dir)

    file := path.Join(out_dir, "pipe_out.txt")
    out_fh, err := os.Create(file)
    if err != nil {
        t.Errorf("couldn't create test file %q: %s", file, err)
        return
    }

    cmds := [][]string{
        []string{"cut", "-f2-"},
        []string{"sort"},
    }

    input_str := "2\ttwo\n3\tthree\n1\tone\n"
    expected_output := "one\nthree\ntwo\n"

    wc, err := fileutil.OpenPipesToWriter(out_fh, cmds)
    if err !=  nil {
        t.Errorf("couldn't OpenPipesToWriter go file %q: %s", file, err)
        return
    }

    fmt.Fprintf(wc, "%s", input_str)
    if err = wc.Close(); err != nil {
        t.Errorf("close on pipes writer failed: %s", err)
        out_fh.Close()
        return
    }

    if err = out_fh.Close(); err != nil {
        t.Errorf("close on file handle failed: %s", err)
        return
    }

    in_fh, err := os.Open(file)
    if err != nil {
        t.Errorf("failed to open input file %q: %s", file, err)
        return
    }

    got_bytes, err := ioutil.ReadAll(in_fh)
    if err != nil {
        t.Errorf("failed to ReadAll from input file %q: %s", file, err)
        return
    }
    in_fh.Close()

    if string(got_bytes) != expected_output {
        t.Errorf("got %q, expected %q", string(got_bytes), expected_output)
        return
    }
}
