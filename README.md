
# FogSync Client

## Overview

This is a client for fogsync, which provides secure file sync across multiple
systems. 

## Install 

Instructions tested on Ubuntu 12.04 

1. Install some basic development packages
  - sudo apt-get install git mercurial ruby bundler

2. Install Go
  - Get Godeb, unpack it (http://blog.labix.org/2013/06/15/in-flight-deb-packages-of-go)
  - ./godeb install 1.4

3. Set up gopath
  - mkdir ~/.local/go
  - echo "export GOPATH=~/.local/go" >> ~/.bashrc
  - . ~/.bashrc

3. Clone fogsync
  - git clone https://github.com/fogcloud/fogsync.git

4. Get build prereqs
  - cd fogsync/src
  - make prereqs

5. Build FogSync
  - make

6. Run FogSync
  - bin/fogsync

7. The client settings form will pop up.
  - Enter your cloud login information.
  - If this is your first machine with FogSync, use the randomly generated
    master key and print the settings page once you have saved it.
  - If you already have machines set up with FogSync, enter your existing
    master key.

## License

FogSync Client is copyright &copy;2014 Nat Tuck. You may copy, modify, and
redistribute it under the terms of the GNU GPL, either version 3 or any later
version as published by the FSF.
