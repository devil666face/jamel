#!/bin/bash
for i in {0..20}; do
  ./bin/jamel-admin -force ubuntu:latest &
done
