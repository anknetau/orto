# Orto

<img src="assets/orto-git.png" alt="Orto logo" width="512" height="100"/>

**Orto** is an orthogonal data manager for Git.  

It provides a clean and independent layer for managing data alongside your Git repositories, keeping your workflows modular and uncluttered.

# What? Why?

Unlike `git stash`, which keeps temporary changes inside the .git internals, Orto acts as a **safe stash**. It lets you save a snapshot of your uncommitted changes outside of the repository, manage them orthogonally, and bring them back when you need them. This keeps your repository history and stash list uncluttered while giving you fine-grained control over your work-in-progress.

## Planned Features

- Written in go for speed
- Orthogonal design: manage data without polluting your Git history  
- Lightweight integration with existing Git workflows  
- Several output options (zip, tar.bz2, raw directories, patch, etc.)
- Include files that git ignores
- Encryption to safely store untracked files with credentials
- Simple setup and usage  

