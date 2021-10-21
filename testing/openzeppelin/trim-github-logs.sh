#!/bin/sh
# removes first 29 characters from file
sed -i -r 's/.{29}//' truffle-expected-output.json
