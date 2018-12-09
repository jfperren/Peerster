
rm -rf submission/*

# mkdir ../tmp
cp -rf . ../tmp
cd ../tmp
rm -rf submission
rm -rf _SharedFiles/*
rm -rf _Downloads/*
rm -rf logs/*

rm Peerster
rm client/client

cd ../Peerster

name="Peerster-jfperren-v3.0.0"

mkdir submission/github.com
mkdir submission/github.com/jfperren

mv ../tmp submission/github.com/jfperren
cd submission/github.com/jfperren/
mv tmp Peerster
cd ../..

tar -zcvf $name.tar.gz github.com
rm -rf github.com
