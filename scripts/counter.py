import tkinter as tk
import sys
from tkinter import IntVar
import threading
import time
from datetime import datetime

window = tk.Tk()

window.title("People Counter")

people_count = IntVar(window, int(sys.argv[1]))

# Show the number of people
people = tk.Label(textvariable=people_count)
people.pack()

# Increase and decrease the number of people
def increase_people():
    people_count.set(int(people_count.get()) + 1)

def decrease_people():
    people_count.set(int(people_count.get()) - 1)

add = tk.Button(
    text="+ Person",
    width=25,
    height=5,
    bg="black",
    fg="black",
    command=increase_people
)

delete = tk.Button(
    text="- Person",
    width=25,
    height=5,
    bg="black",
    fg="black",
    command=decrease_people
)

add.pack()
delete.pack()


def save_people_count():

    while True:
        with open("people_count.txt", "a") as f:
            to_write = datetime.now().isoformat() + "," + str(people_count.get()) + "\n"
            print(to_write)
            f.write(to_write)

        # sleep till the start of the next minute
        time.sleep(10-time.time()%10)

if __name__ == "__main__":
    x = threading.Thread(target=save_people_count)
    x.start()
    window.mainloop()