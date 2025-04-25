This is a dockerfile containing eofparsers, plus the comparer from goevmlab

Example on how to run it:
```
yes "ef0000112233" | docker run -i holiman/omni-eof binaries.txt
```
It can also be run with source- and sink- containers.

Start the sink:
```
docker -v /tmp/foo:/tmpfuzz run holiman/omni-eof-sink src=/fuzztmp/tests.ipc binaries.txt
```
Then start the source (generator):
```
docker -v /tmp/foo:/tmpfuzz run holiman/omni-eof-source
```
