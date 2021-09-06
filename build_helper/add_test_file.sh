find . -name "*.go" | grep -vE "proto|docs|library/util" | while read fname; do
    fpath=`dirname $fname`
    firstline=$(sed -n -e '/^package/p' $fname)
#    if [ ! -f "$fpath/main_test.go" ]; then
    if ls $fpath/*_test.go 1> /dev/null 2>&1; then
      echo "no need to add main_test.go file in $fpath"
    else
      echo $firstline > $fpath/main_test.go
      echo "import \"testing\"\nfunc Test_Main(t *testing.T) {}" >> $fpath/main_test.go
    fi
done