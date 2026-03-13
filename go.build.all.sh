GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./dist/checkout -ldflags="-w -s" ./checkout
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./dist/clear -ldflags="-w -s" ./clear
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./dist/dockerBuild -ldflags="-w -s" ./dockerBuild
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./dist/dockerPush -ldflags="-w -s" ./dockerPush
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./dist/dockerswarmUpdateVer -ldflags="-w -s" ./dockerswarmUpdateVer
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./dist/gitProxy -ldflags="-w -s" ./gitProxy
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./dist/newBuild -ldflags="-w -s" ./newBuild
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./dist/setup-go -ldflags="-w -s" ./setup-go
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./dist/setup-npm -ldflags="-w -s" ./setup-npm

docker cp ./. FOPS-Build:/var/lib/fops/actions/farseers/FOPS-Actions/v1/
docker cp ./. FOPS-AutoBuild:/var/lib/fops/actions/farseers/FOPS-Actions/v1/
CommidId=$(docker ps | grep fops.1 | awk '{print $1}') && docker cp ./. $CommidId:/var/lib/fops/actions/farseers/FOPS-Actions/v1/