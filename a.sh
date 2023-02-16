#!/bin/bash
trap "rm server;kill 0" EXIT

go build -o server
./server -port=8001 &
./server -port=8002 &
./server -port=8003 &

sleep 2
echo ">>> start test"
grpcurl -plaintext -d '{"group":"scores", "key": "Tom"}' 127.0.0.1:8001 pb.GroupCache/Get &
grpcurl -plaintext -d '{"group":"scores", "key": "Tom1"}' 127.0.0.1:8001 pb.GroupCache/Get &
grpcurl -plaintext -d '{"group":"scores", "key": "Tom2"}' 127.0.0.1:8001 pb.GroupCache/Get &
grpcurl -plaintext -d '{"group":"scores", "key": "Tom"}' 127.0.0.1:8002 pb.GroupCache/Get &
grpcurl -plaintext -d '{"group":"scores", "key": "Tom1"}' 127.0.0.1:8002 pb.GroupCache/Get &
grpcurl -plaintext -d '{"group":"scores", "key": "Tom2"}' 127.0.0.1:8002 pb.GroupCache/Get &
grpcurl -plaintext -d '{"group":"scores", "key": "Tom"}' 127.0.0.1:8003 pb.GroupCache/Get &
grpcurl -plaintext -d '{"group":"scores", "key": "Tom1"}' 127.0.0.1:8003 pb.GroupCache/Get &
grpcurl -plaintext -d '{"group":"scores", "key": "Tom2"}' 127.0.0.1:8003 pb.GroupCache/Get &

wait