# GClone - Git Repository Cloning Tool

GClone is a command-line tool that helps you manage Git repository clones with multiple profiles. It allows you to define different SSH hosts and Git configurations for different Git accounts (e.g., personal, work), and automatically applies the appropriate settings when cloning repositories.

![demo](/demo/gclone-demo.gif)

## Features

- Clone repositories using different SSH configurations based on profiles
- Automatically applies Git configurations per profile (username, email, etc.)
- Supports different profile configurations for work, personal, and other accounts
- Colorful and easy-to-use CLI interface with a dedicated UI package

## Installation

```bash
# Clone the repository
git clone git@github.com:user-cube/gclone.git

# Build and install
cd gclone
go install
```

## Usage

### Initialize Configuration

```bash
# Initialize the default configuration
gclone init
```

This creates a configuration file at `~/.gclone/config.yml` with some sample profiles.

### Manage Profiles

```bash
# List all profiles
gclone profile list

# Add a new profile
gclone profile add personal --ssh-host=git-personal --git-username="Your Name" --git-email="your.email@example.com"

# Add a profile with URL patterns for automatic detection
gclone profile add personal --ssh-host=git-personal --git-username="Your Name" --git-email="your.email@example.com" --url-pattern="github.com/your-username" --url-pattern="gitlab.com/your-username"

# Remove a profile
gclone profile remove personal
```

### Clone Repositories

```bash
# Clone a repository using a specific profile
gclone clone git@github.com:user/repo.git --profile=personal

# Clone with automatic profile detection based on URL patterns
gclone clone git@github.com:your-personal-username/repo.git
# GClone will automatically use the personal profile if the URL matches a configured pattern

# Clone with additional options
gclone clone git@gitlab.com:user/repo.git my-repo --profile=work --depth=1 --branch=main
```

> **Note:** GClone only supports SSH URLs (git@github.com:user/repo.git format). HTTP/HTTPS URLs are not supported.

### View Configuration

```bash
# View the current configuration
gclone config
```

## How It Works

When you clone a repository with GClone, it:

1. Reads your profile configuration from `~/.gclone/config.yml`
2. Attempts to automatically detect the appropriate profile based on URL patterns
3. Transforms the repository URL to use the specified SSH host for that profile
4. Clones the repository using the modified URL
5. Applies any Git configurations specified in the profile to the cloned repository

For example, if you have a profile named `personal` with an SSH host of `git-personal`, when you clone `git@github.com:user/repo.git`, GClone will automatically change it to `git@git-personal:user/repo.git`.

> **Note:** GClone only works with SSH URLs in the format `git@github.com:user/repo.git`.

### URL Pattern Matching

GClone can automatically select the appropriate profile based on URL patterns. For example:

- If you have a personal profile with a URL pattern of `github.com/your-username`
- When you run `gclone clone git@github.com:your-username/repo.git`
- GClone will automatically use your personal profile without you needing to specify `--profile=personal`

## Configuration

The configuration file is stored at `~/.gclone/config.yml` and has the following structure:

```yaml
profiles:
  personal:
    name: Personal
    ssh_host: git-personal
    url_patterns:
      - github.com/your-personal-username
      - github.com:your-personal-username
    git_configs:
      user.name: Your Name
      user.email: your.email@example.com
  work:
    name: Work
    ssh_host: git-work
    url_patterns:
      - github.com/your-work-organization
      - github.com:your-work-organization
    git_configs:
      user.name: Your Work Name
      user.email: your.work.email@example.com
```

## SSH Configuration

GClone can help you manage your SSH configurations automatically. For each profile, GClone can generate and maintain the necessary SSH host configuration.

### Using the SSH Config Command

GClone provides a built-in command to create or update SSH configurations:

```bash
# Generate SSH config for a specific profile
gclone ssh-config personal

# Or run without arguments to select from available profiles
gclone ssh-config
```

This command will:

1. Create a `~/.gclone/ssh_config` file with the SSH host configuration for your profile
2. Add an `Include ~/.gclone/ssh_config` directive to your main `~/.ssh/config` file if it doesn't exist

The generated configuration will look like this:

```
# personal profile (added by gclone)
Host github.com-personal
  Hostname github.com
  AddKeysToAgent yes
  UseKeychain yes
  IdentityFile ~/.ssh/github_personal
```

### Manual Configuration

If you prefer to set up SSH configurations manually, you can add them directly to your `~/.ssh/config` file:

```
# Personal GitHub account
Host github.com-personal
  Hostname github.com
  AddKeysToAgent yes
  UseKeychain yes
  IdentityFile ~/.ssh/github_personal

# Work GitHub account
Host github.com-work
  Hostname github.com
  AddKeysToAgent yes
  UseKeychain yes
  IdentityFile ~/.ssh/github_work
```

In this example:

- `Host github.com-personal` defines the alias that matches the `ssh_host` in your GClone profile
- `Hostname github.com` specifies the actual Git server to connect to
- `IdentityFile ~/.ssh/github_personal` specifies which SSH key to use for this host

When GClone transforms a URL from `git@github.com:user/repo.git` to `git@github.com-personal:user/repo.git`, SSH will use the configuration defined for the `github.com-personal` host.
