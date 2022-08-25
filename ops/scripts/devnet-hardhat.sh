#!/bin/sh
cpu_struct=`arch`;
echo $cpu_struct;

node --version;
dpid=`docker ps | grep devnet | awk '{print $1}'`;
echo "Checking for existing docker containers for devnet..."
if [ -z "$dpid" ]
then
    echo "No docker devnet container running...";
else
    docker kill $dpid;
fi

devnet_image=`docker ps -a | grep devnet_local | awk '{print $1}'`
if [ -z "$devnet_image" ]
then
    echo "No docker devnet imagse found...";
else
    docker rm $devnet_image;
fi

echo "Checking CPU structure..."
if [[ $cpu_struct == *"arm"* ]]
then
    echo "Starting arm devnet container..."
    docker run -p 5050:5050 -p 8545:8545 -d --name devnet_local shardlabs/starknet-devnet:0.2.9-arm;
else
    echo "Starting i386 devnet container..."
    docker run -p 5050:5050 -p 8545:8545 -d --name devnet_local shardlabs/starknet-devnet:0.2.9;
fi

echo "Checking for running hardhat process..."

hardhat_image=`docker image ls | grep hardhat | awk '{print $3}'`
if [ -z "$hardhat_image" ]
then
    echo "No docker hardhat image found...";
else
    docker rm $hardhat_image;
fi

dpid=`docker ps | grep hardhat | awk '{print $1}'`;
echo "Checking for existing docker containers for hardhat..."
if [ -z "$dpid" ]
then
    echo "No docker hardhat container running...";
else
    docker kill $dpid;
fi
echo "Starting hardhat..."
docker run --net container:devnet_local -d ethereumoptimism/hardhat