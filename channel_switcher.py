import os
import time
chans = [1,6,11]
wait = 1
i = 0
while True:
    os.system('iw dev wlan1mon set channel {}'.format(chans[i]))
    i = (i + 1) % len(chans)
    time.sleep(wait)