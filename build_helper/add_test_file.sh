find . -name "*.go" | grep -vE "proto|docs|library/util" | while read fname; do
    fpath=`dirname $fname`
    firstline=$(head -n 1 $fname)
    if [ ! -f "$fpath/main_test.go" ]; then
      echo $firstline > $fpath/main_test.go
    fi
done