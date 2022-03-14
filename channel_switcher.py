import os
import time
import sys
two = [1,6,11, 2,7,12, 3,9,13, 4,10,5,8]
five = [36, 40, 44, 48, 52, 56, 60, 64, 100, 104, 108, 112, 116, 120, 124, 128, 132, 136, 140]
chans = two
wait = 0.1
i = 0
wlan = sys.argv[1]
while True:
    os.system('iw dev {} set channel {}'.format(wlan, chans[i]))
    i = (i + 1) % len(chans)
    time.sleep(wait)