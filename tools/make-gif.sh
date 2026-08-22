#!/usr/bin/env bash
# Turn a screen recording into a portfolio demo GIF + poster frame.
#
#   ./tools/make-gif.sh gaming-zone ~/Desktop/rec.mov [start] [duration]
#
# Writes media/<slug>/demo.gif and media/<slug>/poster.jpg, then tells you
# to point projects.json at them.
set -euo pipefail

slug=${1:?usage: make-gif.sh <slug> <recording> [start] [duration]}
src=${2:?missing recording path}
start=${3:-0}
dur=${4:-20}
out="media/$slug"
mkdir -p "$out"

# 900px wide is plenty at portfolio size; 15fps keeps a 20s loop under ~4MB.
# ponytail: two-pass palette. One-pass gifs band badly on dark UI.
pal=$(mktemp -t pal).png
trap 'rm -f "$pal"' EXIT
filters="fps=15,scale=900:-2:flags=lanczos"

ffmpeg -v error -ss "$start" -t "$dur" -i "$src" -vf "$filters,palettegen=stats_mode=diff" -y "$pal"
ffmpeg -v error -ss "$start" -t "$dur" -i "$src" -i "$pal" \
  -lavfi "$filters[x];[x][1:v]paletteuse=dither=bayer:bayer_scale=3" \
  -loop 0 -y "$out/demo.gif"

# Poster = the first frame. Shown until the demo scrolls into view.
ffmpeg -v error -ss "$start" -i "$src" -vframes 1 -vf "scale=900:-2" -q:v 4 -y "$out/poster.jpg"

size=$(du -h "$out/demo.gif" | cut -f1)
echo "✅ $out/demo.gif ($size)  +  $out/poster.jpg"
[ "${size%M}" != "$size" ] && awk -v s="${size%M}" 'BEGIN{if(s+0>5) print "⚠️  over 5MB — rerun with a shorter duration or fps=12"}'
echo
echo "Now set in projects.json for \"$slug\":"
echo "  \"media\": { \"aspect\": \"phone|desktop\", \"demo\": \"$out/demo.gif\", \"poster\": \"$out/poster.jpg\" }"
