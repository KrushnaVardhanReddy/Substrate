# Substrate Launch Video Production Guide

This document outlines the end-to-end workflow for creating a professional, AI-narrated product launch video for Substrate using a 100% open-source, local Linux stack leveraging an NVIDIA GPU.

## 🛠 Step 1: The Setup (Install the Toolchain)
Install the "Holy Trinity" of open-source video production on Linux.

1. **Install OBS Studio (Screen Recording):**
   ```bash
   sudo apt install obs-studio
   ```
2. **Install Kdenlive (Video Editing):**
   ```bash
   sudo apt install kdenlive
   ```
3. **Setup Coqui XTTS v2 (AI Voice Generation):**
   Create a virtual environment for the AI audio generator so it doesn't conflict with project dependencies:
   ```bash
   python3 -m venv ~/video-tts-env
   source ~/video-tts-env/bin/activate
   pip install TTS
   ```

## 🎙 Step 2: Generate the AI Audio (Do this FIRST)
*Pro Tip: Always generate the audio before you record your screen. It is 100x easier to click around the app to match an audio track than it is to stretch/cut audio to match a video.*

1. **Write the Script in chunks:** Don't try to generate a 3-minute audio file at once. Break it into paragraphs (e.g., `01_intro.txt`, `02_demo.txt`, `03_enterprise.txt`).
2. **Find a Voice Clone Sample:** Find a 5-second YouTube video of a tech presenter you like. Download it and save it as `speaker.wav`. (Ensure there is no background music in the clip).
3. **Generate the Narration:**
   Run this command for each script chunk. Your GPU will process it almost instantly:
   ```bash
   tts --text "We've all been there. You merge a simple schema change, and suddenly three downstream teams are broken in production. Meet Substrate." \
       --model_name tts_models/multilingual/multi-dataset/xtts_v2 \
       --speaker_wav speaker.wav \
       --language_idx en \
       --out_path 01_intro.wav
   ```
4. Listen to `01_intro.wav`. If the pacing is off, adjust the text (add commas or periods for pauses) and re-run.

## 🎥 Step 3: Screen Recording in OBS
With the audio files ready, record the visual components.

1. **Configure OBS for High-Quality Text:**
   * Go to **Settings > Output > Recording**.
   * Set Recording Format to `mkv` (prevents file corruption if OBS crashes).
   * Set Video Encoder to **NVIDIA NVENC H.264**.
   * Set Rate Control to **CQP** and the CQ Level to `15` (Guarantees crystal clear terminal/code text with no compression artifacts).
   * Go to **Settings > Video** and set Base and Output resolution to `1920x1080` at `60 FPS`.

2. **The "Puppet" Recording Technique:**
   * Open the Substrate dashboard, VS Code, or GitHub.
   * Play `01_intro.wav` out loud on your speakers.
   * Hit **Record** on OBS.
   * As the AI voice speaks, move your mouse, type code, or click through the dashboard *in exact rhythm with the voice*. 
   * Stop the recording. You now have a video clip perfectly synced to your audio.

## 🎞 Step 4: The Edit (Kdenlive)
1. Open **Kdenlive**.
2. Drag your generated AI `.wav` files into the bottom audio track in sequential order.
3. Drag your OBS video recordings into the video track above the audio.
4. Because of the "Puppet" technique, the video should already be ~95% synced. Trim the edges as needed.
5. **Zoom & Pan (Crucial for Dev Tools):**
   * Viewers will watch on mobile devices. You *must* zoom in on the code or the PR comment.
   * Select a clip, apply the **"Transform"** effect.
   * Add keyframes to zoom in on the GitHub PR comment exactly when the AI voice says: *"Substrate instantly calculates the blast radius..."*
6. **Add Background Music:**
   * Download a subtle, upbeat electronic background track (e.g., from YouTube Audio Library).
   * Place it on Audio Track 2. 
   * Lower the volume to `-20dB` or `-25dB` so it does not overpower the AI narration.

## 🚀 Step 5: Export and Launch
1. In Kdenlive, click **Project > Render**.
2. Choose **MP4-H264/AAC**.
3. Check the "More Options" box and ensure the Video Quality slider is pushed to maximum quality.
4. Hit **Render to File**.

*You now have a polished, GPU-accelerated, AI-narrated product launch video ready for YouTube, Twitter, and Product Hunt!*
