<p align="center">
<pre>
> ██████╗  █████╗  ██████╗████████╗ ██████╗ 
> ██╔══██╗██╔══██╗██╔════╝╚══██╔══╝██╔═══██╗
> ██████╔╝███████║██║        ██║   ██║   ██║
> ██╔═══╝ ██╔══██║██║        ██║   ██║   ██║
> ██║     ██║  ██║╚██████╗   ██║   ╚██████╔╝
> ╚═╝     ╚═╝  ╚═╝ ╚═════╝   ╚═╝    ╚═════╝ 
> Pressure → Aim → Constraint → Terminate → Outcome reflection
</pre>
</p>

**A focused work timer built on psychology, not just pomodoros.**

PACTO replaces the vague “did you work?” with an honest “did you produce what you planned?” It does this by combining a fixed deadline, an explicit outcome, a mandatory hard stop, and a forced review—turning well-studied cognitive principles into a simple, repeatable loop.

## How it works

1. **Decide** what concrete output you’ll create in the next bout.
2. **Start** the timer. The window minimizes—no distractions.
3. **Stop** the moment the timer ends, even if you’re mid-flow.
4. **Review**: Complete, shift the outcome forward, or discard it.

Every session is logged locally. No accounts, no cloud.

## The psychology behind the loop

PACTO’s workflow isn’t arbitrary. It’s a deliberate chain of evidence-based techniques:

- **Fixed deadlines fight Parkinson’s Law.** Work expands to fill the time you give it; a hard limit creates the healthy urgency that defeats perfectionism.
- **Naming an outcome is an implementation intention.** Specifying *what* and *when* automates focus so you don’t rely on willpower alone.
- **A hard stop exploits the Zeigarnik Effect.** Interrupted tasks stick in memory, creating mental tension that pulls you into the next bout—especially when you *shift* an unfinished outcome.
- **Forced review builds metacognition.** Explicitly judging your output (complete / shift / discard) creates a feedback loop that improves planning and performance over time, much like self-regulated learning interventions.

Together, these form **P**ressure → **A**im → **C**onstraint → **T**erminate → **O**utcome reflection.

## Features

- Configurable bout length (1–180 minutes)
- Outcome commitment before each timer starts
- Shift mechanic to carry unfinished outcomes forward
- Session history grouped by date with time ranges
- Start and end chimes
- Optional lo‑fi focus music (drop `.mp3` files next to the `.exe`)
- Theme switcher (Windows 7, Windows XP, Arcade)
- Single portable `.exe` — no install required

---

## Download

Get the latest `.exe` from [Releases](https://github.com/Tushar-R-Tyagi/pacto-desktop/releases).

---

## Build from source

```bash
git clone https://github.com/Tushar-R-Tyagi/pacto-desktop.git
cd pacto-desktop
wails build
