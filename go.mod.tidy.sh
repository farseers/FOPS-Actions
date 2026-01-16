mods=$(find . -name "go.mod")
curDir=$(pwd)
for file in $mods
do
  dir=$(dirname $file)
  echo $dir
  cd $dir
  go get -u && go mod tidy
  cd $curDir
done