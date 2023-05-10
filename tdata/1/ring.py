#!/usr/bin/env python3

import sys
import time
from PyQt4.QtCore import Qt, QTime
from PyQt4.QtGui import (
    QApplication, QDialog, QHBoxLayout, QLabel, QPushButton, QVBoxLayout,)


if len(sys.argv) < 2:
    print("usage: ring.py HH[:MM] <optional comment>")
    sys.exit()

start = list(map(int, sys.argv[1].split(":")))
if len(start) == 1:
    start.append(0)
comment = "Wake Up!"
if len(sys.argv) > 2:
    comment = " ".join(sys.argv[2:])

due = QTime(start[0], start[1])
while QTime.currentTime() < due:
    time.sleep(20)

app = QApplication(sys.argv)
font = app.font()
font.setPointSize(14)
app.setFont(font)
form = QDialog()
form.setWindowFlags(Qt.WindowStaysOnTopHint)
form.setWindowTitle("Ring!")
label = QLabel("<p>&nbsp;</p><p><b>%s</b></p><p>&nbsp;</p>" % comment)
label.setAlignment(Qt.AlignCenter)
font = label.font()
font.setPointSize(font.pointSize() * 5)
label.setFont(font)
button = QPushButton("OK")
button.setDefault(True)
hbox = QHBoxLayout()
hbox.addStretch()
hbox.addWidget(button)
hbox.addStretch()
form.resize(200, 200)
layout = QVBoxLayout()
layout.addWidget(label)
layout.addLayout(hbox)
form.setLayout(layout)
button.clicked.connect(form.accept)
form.move(QApplication.desktop().width() / 2,
          QApplication.desktop().height() / 2)
form.show()
app.exec_()
