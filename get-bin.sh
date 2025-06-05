#!/bin/sh
URL=$(curl -s https://api.github.com/repos/cxpsemea/cx1e2e/releases/latest | grep "https.*cx1e2e\"" | sed "s/^.* \"//g" | sed "s/\".*$//g")
echo URL is $URL
wget $URL -O cx1e2e-bin
chmod +x ./cx1e2e-bin
