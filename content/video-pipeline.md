
## Problem

A large MP4 library is not a streaming library. A phone on a weak connection cannot pull a 2 GB file, players will not seek reliably inside a monolithic mp4, and a codec that works on one device stalls on another. The fix is well known — segment each video, encode a ladder of qualities, and publish a manifest the player can switch inside — but doing it by hand means a bespoke FFmpeg invocation per file.

At batch scale the encode stops being the hard part. The hard parts are the ones that only show up at file 400: a run that dies overnight with no way to tell which outputs are safe, an FFmpeg command that exits `0` while writing a playlist that references segments it never wrote, and source files deleted after a "successful" conversion nobody validated.

I built tooling that treats encoding as a pipeline stage rather than a command — one that reports what it did, checks its own output, and can be stopped without corrupting anything.

## What I built

**HLS transcoder.** Points at a folder of folders and converts each one to adaptive-bitrate HLS. Probes the source, picks the quality ladder, encodes video and audio as separate renditions, writes the master playlist, validates the result, optionally packages it as a ZIP and removes the source. Runs as a config-driven CLI or a small desktop GUI for non-technical operators.

**GPU-accelerated variant.** Same pipeline with hardware encode: it enumerates each NVIDIA GPU's real capabilities — NVENC support per codec, NVDEC, free VRAM, concurrent session limit — and on a two-GPU box runs the H.264 ladder on one card while the alternative codec encodes on the other. Also emits hover-preview thumbnails and a short preview clip per video.

**Scene fragmenter.** Reduces a video to one representative still per shot, for feeding frames to an image model and recomposing them. A 20-minute clip at 24 fps is ~28,000 frames and almost all of them are duplicates; this keeps only the frames where the picture actually changes, with a manifest tying each image back to its source timestamp.

**Subtitle extractor.** Batch-pulls English and Spanish subtitle tracks out of a library, including image-based tracks that have to be OCR'd, and validates every `.srt` it produces before letting the pipeline delete anything.

## How it was done

**Audio is encoded once, not once per rendition.** Video renditions are encoded `-an` into per-quality folders; audio goes to its own `aac.m3u8` and is attached to the master playlist as an `EXT-X-MEDIA` audio group. Five video renditions therefore ship one audio track, not five copies of it.

**The ladder is derived from the source, not assumed.** `ffprobe` reads dimensions, bitrate and duration; the detector maps height to a source tier and emits only profiles at or below it, so a 360p source never gets upscaled into a fake 720p rendition. Ladder: H.264 at 2077k/720p and 800k/360p, VP9 at 1291k/720p, 768k/480p and 594k/360p — two codecs so older devices get H.264 while VP9 clients get the smaller files.

**Fragmented MP4 with independent segments.** `hls_segment_type fmp4`, `independent_segments`, VOD playlist type, configurable segment duration (default 5s). One consequence had to be handled explicitly: when FFmpeg does not emit an `init.mp4`, the encoder synthesises one with a zero-length `frag_keyframe+empty_moov+default_base_moof` pass rather than shipping a rendition no player can start.

**The master playlist is written by hand.** `BANDWIDTH`, `RESOLUTION` and exact `CODECS` strings per rendition (`avc1.64001f`, `vp09.00.31.08.00.01.01.01.00`, …), sorted descending, H.264 group before VP9. Players make their first pick from this file alone, so it is the one place where being approximately right is worse than useless.

**Validation is a separate stage from encoding.** A conversion that exits cleanly is still checked: playlist readable, init segment present and non-empty, every segment referenced by the playlist actually on disk, and a real `ffmpeg -f null` decode of the playlist end to end. Only output that passes is eligible for the "delete the source" path — the destructive step is gated on the check, not on the exit code.

**Stopping is cooperative, not fatal.** `SIGINT`/`SIGTERM` set a flag; the run finishes the video in flight and then exits at a folder boundary. Killing FFmpeg mid-file would leave a half-written rendition that looks complete on disk, which is exactly the failure the validator exists to prevent.

**Scene detection is deterministic, and that was the point.** FFmpeg's own `select='gt(scene,T)',showinfo` filter scores each frame against the previous one; timestamps are parsed back out of the log so detection and extraction happen in a single decode pass. An ML shot detector would be slower and would not return the same cuts twice — a fixed threshold is reproducible and tunable with one number. The detector sits behind an interface so a smarter backend can replace it without touching the extractor.

**One PGS decoder mistake worth keeping.** Image-based subtitles were first handled by hand-rolling the PGS RLE and palette decode. It was subtly wrong — some events rendered with a neighbouring event's glyphs, producing text that OCR'd cleanly and was simply the wrong dialogue. Plausible-looking bad output is worse than a crash. It was replaced by letting FFmpeg's own decoder burn the subtitle stream onto a blackened copy of the video, grabbing the frame at each event's midpoint in one sequential pass, and OCR'ing that.

**Batch runs are manifest-driven.** The run writes the list of expected outputs up front, then the final pass walks that manifest and decides per job: pass keeps output and deletes input, fail deletes output and keeps input, skip touches nothing. Deletion happens once, at the end, from the manifest, and `--dry-run` prints every decision without acting. OCR is the bottleneck, so jobs are the unit of parallelism across a process pool.

## Stack

Python 3.10+, FFmpeg / FFprobe, tkinter, NVIDIA NVENC + NVDEC (`nvidia-smi` capability probe), Tesseract OCR with `eng`/`spa` data, Pillow, `ProcessPoolExecutor`, zipfile.

