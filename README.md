# Trkr
## Description
Trkr is a simple REST endpoint that can be used to track user activity
It required MongoDB up and running with a database and a collection created

## How to use
./trkr -port <httpPort> -mongoAddr <mongoDB connection string> -mongoDatabase <MongoDB database> -mongoCollection <MongoDB collection>

## How to build for your environment
make build 

## How to build for linux
make build-linux
