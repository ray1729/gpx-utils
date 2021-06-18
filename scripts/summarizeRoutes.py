#!/usr/bin/python3

import csv
import sys

path = sys.argv[1]

with open(path) as f:
    r = csv.reader(f)
    skip = True
    for x in r:
        if skip:
            skip = False
            continue
        coffee = x[7].strip()
        lunch = x[8].strip()
        tea = x[9].strip()
        if coffee and tea:
            print(coffee + ", " + lunch + ", " + tea)


