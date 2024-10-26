set -e

rm -rf .bin
cd src

echo 'building darwin-amd64...'
GOOS=darwin GOARCH=amd64 go build -o ../.bin/darwin-amd64/faroe
echo 'building darwin-arm64...'
GOOS=darwin GOARCH=arm64 go build -o ../.bin/darwin-arm64/faroe

echo 'building linux-amd64...'
GOOS=linux GOARCH=amd64 go build -o ../.bin/linux-amd64/faroe
echo 'building linux-arm64...'
GOOS=linux GOARCH=arm64 go build -o ../.bin/linux-arm64/faroe

echo 'building windows-amd64...'
GOOS=windows GOARCH=amd64 go build -o ../.bin/windows-amd64/faroe
echo 'building windows-arm64...'
GOOS=windows GOARCH=arm64 go build -o ../.bin/windows-arm64/faroe

cd ..
cd .bin
for dir in $(ls -d *); do
    cp ../LICENSE "$dir"/LICENSE
    zip -r "$dir".zip $dir
    rm -rf $dir
done
cd ..

echo 'done!'