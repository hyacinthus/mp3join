# MP3 Joiner (Golang Library)

Pure go library for joining mp3 files to one.

The library will read all files in memery, so it can not process big files.

## Usage

```golang

joiner := mp3join.New()

// readers is the input mp3 files
for reader := range readers {
    err := j.Append(reader)
    if err != nil {
        return err
    }
}

dest := joiner.Reader()

// dest is a mp3 output reader

```
