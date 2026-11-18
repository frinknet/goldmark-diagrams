#!/bin/sh

docker rmi "goldmark-diagrams" 2>/dev/null || true
docker build -t "goldmark-diagrams:test" .
