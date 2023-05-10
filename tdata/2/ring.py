#!/usr/bin/env python3
# Copyright © 2021 Mark Summerfield. All rights reserved.
# License: GPLv3

import datetime
import re
import sys
import time
import tkinter as tk
import tkinter.font
import tkinter.ttk as ttk


def main():
    USAGE = 'usage: ring.py H[H[:M[M]]] <optional message>'
    if len(sys.argv) == 1 or sys.argv[1] in {'-h', '--help'}:
        raise SystemExit(USAGE)
    match = re.fullmatch(r'(?P<h>\d\d?)(?::(?P<m>\d\d?))?', sys.argv[1])
    if match is None:
        raise SystemExit(USAGE)
    when = datetime.time(int(match.group('h')),
                         int(match.group('m') or '0'))
    message = ' '.join(sys.argv[2:]) or None
    while True:
        if datetime.datetime.now().time() >= when:
            break
        time.sleep(20) # secs
    popup(message or
          f'It is {datetime.datetime.now().time().isoformat("minutes")}')


def popup(message):
    def on_close(_event=None):
        app.quit()

    app = tk.Tk()
    app.title(f'Ring!')
    font = tkinter.font.Font(family='Helvetica', size=36, weight='bold')
    window = ttk.Label(app, text=message, padding='5m', font=font,
                       foreground='red', background='cornsilk')
    window.pack()
    app.bind(f'<Control-q>', on_close)
    app.bind(f'<Escape>', on_close)
    app.protocol('WM_DELETE_WINDOW', on_close)
    app.mainloop()


if __name__ == '__main__':
    main()
