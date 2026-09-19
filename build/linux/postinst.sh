#!/bin/sh
# Refresh the icon cache; not every desktop-less system ships the tool.
gtk-update-icon-cache -f -t /usr/share/icons/hicolor || true
