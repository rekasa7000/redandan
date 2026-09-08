# Overview — What is Reliva?

## The Problem

You have multiple jobs. You have a personal life. You have passwords to remember, deadlines to hit, payroll dates to track, and vacations to look forward to. Each of these lives in a different app, a different tab, a different mental bucket — and managing all of them takes energy that should be spent on the actual work.

The friction is real:
- You forget which task belongs to which job
- You can't remember a password and have to reset it
- A deadline sneaks up on you because it was in a different calendar
- You open 5 different apps just to know what today looks like

## The Solution

Reliva is a **personal operating system** — a single, private system that holds everything that matters to you, surfaces what's relevant right now, and gets out of the way.

It is not a product for everyone. It is engineered for one person (you) and optimized for your specific friction points.

## Core Philosophy

**Reduce cognitive load by centralizing everything into one trusted system.**

This means:
- One place to see all your tasks across all jobs and life areas
- One place to check your upcoming week (deadlines, payroll, travel)
- One place to retrieve any password — without browser-vendor lock-in
- Notifications that come to you, not the other way around

## What Reliva Is Not

- It is not a project management tool for teams
- It is not a public product (yet)
- It is not trying to replace your calendar app or notes app entirely
- It does not need to scale to thousands of users — it scales to one

## The Integration Mindset

Work and personal life are not opposites to be "balanced." They are both part of your one life. Reliva treats them that way — tasks from Job 1 and a vacation trip live in the same system, displayed in the same timeline, notified through the same channel.

You don't need to context-switch between apps. You check Reliva.

## The Long-Term Potential

While this is built as a personal tool, the architecture is clean enough to evolve. If it reaches a point where it could benefit others, the foundation is already there:

- Proper auth system (not hardcoded credentials)
- API-first design
- Multi-platform from day one
- Clear separation of concerns

But that's not the current objective. The current objective is: **make your daily life smoother, starting with tasks and passwords.**

## Modules at a Glance

| Module | Purpose |
|---|---|
| Task Tracker | All your tasks across jobs and life, in one view |
| Calendar & Events | Deadlines, payroll, vacations — unified timeline |
| Password Vault | Encrypted credentials, auto-fill in browser |
| Notifications | Proactive alerts that come to you |
| Dashboard | What matters today — at a glance |

## Platforms at a Glance

| Platform | Why |
|---|---|
| Web (Vercel) | Primary interface, accessible anywhere |
| Mobile (Capacitor) | Check tasks and get push notifications on your phone |
| Browser Extension | Password autofill and quick task-add without opening a tab |
