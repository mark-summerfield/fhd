#!/usr/bin/env python3
# Copyright © 2021 Mark Summerfield. All rights reserved.
# License: GPLv3

import datetime
import re
import sys
import time
import tkinter as tk
import tkinter.font


def main():
    USAGE = 'usage: ring.py H[H[:M[M]]] <optional message>'
    if len(sys.argv) == 1 or sys.argv[1] in {'-h', '--help'}:
        raise SystemExit(USAGE)
    match = re.fullmatch(r'(?P<h>\d\d?)(?::(?P<m>\d\d?))?', sys.argv[1])
    if match is None:
        raise SystemExit(USAGE)
    when = datetime.time(int(match.group('h')),
                         int(match.group('m') or '0'))
    message = ' '.join(sys.argv[2:])
    while True:
        if datetime.datetime.now().time() >= when:
            break
        time.sleep(20) # secs
    popup(message.title() if message else
          f'It\'s {datetime.datetime.now().time().isoformat("minutes")}')


def popup(message):
    def on_close(_event=None):
        app.quit()

    app = tk.Tk()
    app.title('Ring!')
    font = tkinter.font.Font(family='Helvetica', size=48, weight='bold')
    window = tk.Button(app, text=message, padx='5m', pady='5m', font=font,
                       foreground='red', background='cornsilk',
                       command=app.quit)
    window.pack()
    app.bind(f'<Control-q>', on_close)
    app.bind(f'<Escape>', on_close)
    app.protocol('WM_DELETE_WINDOW', on_close)
    app.wait_visibility(app)
    app.wm_attributes('-topmost', True)
    app.mainloop()


if __name__ == '__main__':
    main()
