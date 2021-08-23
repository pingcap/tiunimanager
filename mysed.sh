#!/bin/bash


allfiles=`find "$(cd .; pwd)" | grep .go`

for file in $allfiles;  
do   
#	echo $file is filename
	sed -i "" "s/github.com\/pingcap-inc\/tiem\/library\/firstparty/github.com\/pingcap\/tiem\/library\/firstparty/g" $file
done   
