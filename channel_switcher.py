import os
import time
chans = [1,6,11, 2,7,12, 3,9,13, 4,10,5,8]
wait = 0.2
i = 0
while True:
    os.system('iw dev wlan1 set channel {}'.format(chans[i]))
    i = (i + 1) % len(chans)
    time.sleep(wait)