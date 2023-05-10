#!/usr/bin/env python3
# Copyright © Qtrac Ltd. 2019-21. All rights reserved.

import collections
import os
import re

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
        self.words = set()
        self.all_expanded = False
        self.make_widgets()
        self.layout_widgets()
        self.make_bindings()
        self.SetSize(400, 760)
        self.read_words()
        self.update_ui()


    def make_widgets(self):
        self.panel = wx.Panel(self)
        self.wordLabel = wx.StaticText(self.panel, label='&Word')
        self.searchText = wx.TextCtrl(self.panel, style=wx.TE_PROCESS_ENTER)
        self.matchesLabel = wx.StaticText(self.panel, label='&Matches')
        self.matchesTree = wx.TreeCtrl(
            self.panel,
            style=wx.TR_HIDE_ROOT | wx.TR_HAS_BUTTONS | wx.TR_SINGLE)
        font = self.GetFont()
        font.PointSize += 2
        self.matchesTree.SetFont(font)
        self.copyButton = wx.Button(self.panel, label='&Copy')
        self.expandButton = wx.Button(self.panel, label='Expand &All')
        self.searchButton = wx.Button(self.panel, label='&Search')
        self.quitButton = wx.Button(self.panel, label='&Quit')
        self.CreateStatusBar()


    def layout_widgets(self):
        PAD = 5
        flag = wx.ALL | wx.ALIGN_CENTER_VERTICAL
        word_row = wx.BoxSizer(wx.HORIZONTAL)
        word_row.Add(self.wordLabel, 0, flag=flag)
        word_row.Add(self.searchText, 1, flag=flag | wx.EXPAND, border=PAD)
        button_row = wx.BoxSizer(wx.HORIZONTAL)
        button_row.Add(self.copyButton, 0, flag=flag, border=PAD)
        button_row.AddStretchSpacer()
        button_row.Add(self.expandButton, 0, flag=flag, border=PAD)
        button_row.Add(self.searchButton, 0, flag=flag, border=PAD)
        button_row.Add(self.quitButton, 0, flag=flag, border=PAD)
        flag |= wx.EXPAND
        sizer = wx.BoxSizer(wx.VERTICAL)
        sizer.Add(word_row, 0, flag=flag, border=PAD)
        sizer.Add(self.matchesLabel, 0, flag=wx.LEFT | wx.RIGHT, border=PAD)
        sizer.Add(self.matchesTree, 1, flag=flag, border=PAD)
        sizer.Add(button_row, 0, flag=flag)
        self.panel.SetSizer(sizer)
        self.panel.Fit()


    def make_bindings(self):
        letters = ('W', 'M') # Mapping wx.Labels to an associated widget
        accels = []
        for i, letter in enumerate(letters):
            accels.append(wx.AcceleratorEntry(wx.ACCEL_ALT, ord(letter), i))
        self.SetAcceleratorTable(wx.AcceleratorTable(accels))
        for i, widget in enumerate((self.searchText, self.matchesTree)):
            self.Bind(wx.EVT_MENU,
                      lambda _event, widget=widget: widget.SetFocus(), id=i)
        self.copyButton.Bind(wx.EVT_BUTTON, self.on_copy)
        self.expandButton.Bind(wx.EVT_BUTTON, self.on_expand)
        self.searchButton.Bind(wx.EVT_BUTTON, self.on_search)
        self.quitButton.Bind(wx.EVT_BUTTON, lambda _event: self.Close())
        self.searchText.Bind(wx.EVT_TEXT_ENTER, self.on_search)
        self.searchText.Bind(wx.EVT_TEXT, self.update_ui)
        self.Bind(wx.EVT_CHAR_HOOK, self.OnKeyUp)


    def on_copy(self, event):
        text = self.matchesTree.GetItemText(self.matchesTree.Selection)
        if text and wx.TheClipboard.Open():
            try:
                wx.TheClipboard.SetData(wx.TextDataObject(text))
            finally:
                wx.TheClipboard.Close()


    def on_expand(self, _event=None):
        selected = self.matchesTree.Selection
        if self.all_expanded:
            self.matchesTree.CollapseAll()
            label = 'Expand &All'
        else:
            self.matchesTree.ExpandAll()
            label = 'Collapse &All'
        self.matchesTree.EnsureVisible(selected)
        self.expandButton.Label = label
        self.all_expanded = not self.all_expanded


    def on_search(self, _event=None):
        if self.searchText.IsEmpty():
            return
        self.StatusBar.StatusText = 'Searching...'
        busy = wx.BusyCursor()
        try:
            regex = re.compile(self.searchText.GetValue().casefold()
                               .replace(' ', '.'))
            matches = [word for word in self.words if regex.fullmatch(word)]
            prefix = os.path.commonprefix(matches)
            self.matchesTree.DeleteAllItems()
            root = self.matchesTree.AddRoot('Matches')
            if not prefix or len(matches) < 27:
                for word in sorted(matches):
                    self.matchesTree.AppendItem(root, word)
            else:
                self._populate_tree_with_groups(matches, prefix, root)
            self._report_matches(len(matches))
        finally:
            del busy


    def _populate_tree_with_groups(self, matches, prefix, root):
        grouped_words = collections.defaultdict(list)
        index = len(prefix)
        for word in matches:
            letter = ' ' if word == prefix else word[index]
            grouped_words[letter].append(word)
        for letter, words in sorted(grouped_words.items()):
            if len(words) == 1:
                parent = root
            else:
                parent = self.matchesTree.AppendItem(root, prefix + letter)
                self.matchesTree.SetItemTextColour(parent, wx.BLUE)
            for word in sorted(words):
                self.matchesTree.AppendItem(parent, word)


    def _report_matches(self, count):
        if count:
            s = '' if count == 1 else 'es'
            message = f'Found {count:,} match{s}'
            self.matchesTree.SetFocus()
            self.matchesTree.SelectItem(self.matchesTree.FirstVisibleItem)
        else:
            message = 'No matches found'
        self.StatusBar.StatusText = message


    def maybe_expand_or_collapse(self, event=None):
        item = self.matchesTree.FocusedItem
        if item:
            if self.matchesTree.IsExpanded(item):
                self.matchesTree.Collapse(item)
            else:
                self.matchesTree.Expand(item)


    def update_ui(self, _event=None):
        self.searchButton.Enable(not self.searchText.IsEmpty())


    def OnKeyUp(self, event):
        keyCode = event.GetKeyCode()
        if keyCode == wx.WXK_SPACE:
            self.maybe_expand_or_collapse()
        elif keyCode == wx.WXK_ESCAPE:
            self.Close()
        event.Skip()


    def read_words(self):
        self.words.clear()
        for word in open(WORDFILE):
            self.words.add(word.strip().casefold())
        self.StatusBar.StatusText = f'Read {len(self.words):,} words'


if __name__ == '__main__':
    main()
