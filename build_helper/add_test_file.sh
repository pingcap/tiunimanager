find . -name "*.go" | grep -vE "proto|docs|library/util" | while read fname; do
    fpath=`dirname $fname`
    firstline=$(sed -n -e '/^package/p' $fname)
    if [ ! -f "$fpath/main_test.go" ]; then
      echo $firstline > $fpath/main_test.go
      echo "\nimport \"testing\"\n\nfunc TestMain(m *testing.M) {}" >> $fpath/main_test.go
    fi
done