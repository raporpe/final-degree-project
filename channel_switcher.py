import os
import time
two = [1,6,11, 2,7,12, 3,9,13, 4,10,5,8]
five = [36, 40, 44, 48, 52, 56, 60, 64, 100, 104, 108, 112, 116, 120, 124, 128, 132, 136, 140]
chans = two
wait = 0.1
i = 0
while True:
    os.system('iw dev wlan1 set channel {}'.format(chans[i]))
    i = (i + 1) % len(chans)
    time.sleep(wait)