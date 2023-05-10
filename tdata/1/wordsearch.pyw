#!/usr/bin/env python3
# Copyright © Qtrac Ltd. 2019. All rights reserved.

import re
import sys

import wx


WORDFILE = '/usr/share/dict/words'


def main():
    app = wx.App()
    form = Form(None, title="Andrea's Word Search")
    form.Show()
    app.MainLoop()


class Form(wx.Frame):

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.create_widgets()
        self.layout_widgets()
        self.create_bindings()
        self.SetSize(300, 700)
        self.read_words()
        self.update_ui()
        

    def create_widgets(self):
        self.panel = wx.Panel(self)
        self.wordLabel = wx.StaticText(self.panel, label='&Word')
        self.searchText = wx.TextCtrl(self.panel, style=wx.TE_PROCESS_ENTER)
        self.matchesLabel = wx.StaticText(self.panel, label='&Matches')
        self.matchesList = wx.ListBox(self.panel, style=wx.LB_SORT) 
        self.searchButton = wx.Button(self.panel, label='&Search')
        self.quitButton = wx.Button(self.panel, label='&Quit')
        self.CreateStatusBar()


    def layout_widgets(self):
        border = 5
        flag = wx.ALL | wx.ALIGN_CENTER_VERTICAL
        word_row = wx.BoxSizer(wx.HORIZONTAL)
        word_row.Add(self.wordLabel, 0, flag=flag)
        word_row.Add(self.searchText, 1, flag=flag | wx.EXPAND,
                     border=border)
        button_row = wx.BoxSizer(wx.HORIZONTAL)
        button_row.AddStretchSpacer()
        button_row.Add(self.searchButton, 0, flag=flag, border=border)
        button_row.Add(self.quitButton, 0, flag=flag, border=border)
        flag |= wx.EXPAND
        sizer = wx.BoxSizer(wx.VERTICAL)
        sizer.Add(word_row, 0, flag=flag, border=border)
        sizer.Add(self.matchesLabel, 0, flag=wx.LEFT | wx.RIGHT,
                  border=border)
        sizer.Add(self.matchesList, 1, flag=flag, border=border)
        sizer.Add(button_row, 0, flag=flag)
        self.panel.SetSizerAndFit(sizer)


    def create_bindings(self):
        letters = ('W', 'M')
        accels = []
        for i, letter in enumerate(letters):
            accels.append(wx.AcceleratorEntry(wx.ACCEL_ALT, ord(letter), i))
        self.SetAcceleratorTable(wx.AcceleratorTable(accels))
        for i, widget in enumerate((self.searchText, self.matchesList)):
            self.Bind(wx.EVT_MENU,
                      lambda _event, widget=widget: widget.SetFocus(), id=i)
        self.searchButton.Bind(wx.EVT_BUTTON, self.on_search)
        self.quitButton.Bind(wx.EVT_BUTTON, lambda _event: self.Close())
        self.searchText.Bind(wx.EVT_TEXT_ENTER, self.on_search)
        self.searchText.Bind(wx.EVT_TEXT, self.update_ui)


    def on_search(self, _event=None):
        if self.searchText.IsEmpty():
            return
        self.GetStatusBar().SetStatusText('Searching...')
        busy = wx.BusyCursor()
        try:
            self.matchesList.Clear()
            pattern = self.searchText.GetValue().casefold().replace(' ',
                                                                    '.')
            regex = re.compile(pattern)
            for word in self.words:
                if regex.fullmatch(word):
                    self.matchesList.Append(word)
            count = self.matchesList.GetCount()
            if count:
                s = '' if count == 1 else 'es'
                message = f'Found {count:,} match{s}'
            else:
                message = 'No matches found'
            self.GetStatusBar().SetStatusText(message)
        finally:
            del busy


    def update_ui(self, _event=None):
        self.searchButton.Enable(not self.searchText.IsEmpty())


    def read_words(self):
        words = set()
        for word in open(WORDFILE):
            words.add(word.strip().lower())
        self.words = words
        self.GetStatusBar().SetStatusText(f'Read {len(self.words):,} words')


if __name__ == '__main__':
    main()
