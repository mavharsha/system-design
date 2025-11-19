# Deploying Interactive Mindmap to GitHub Pages

This guide explains how the interactive mindmap is automatically deployed to GitHub Pages using GitHub Actions.

## How It Works

### Automatic Deployment

1. **Trigger**: The workflow runs automatically when:
   - You push changes to the `main` branch
   - Changes are made to files in the `mind-map-interactive/` directory
   - You manually trigger it from the Actions tab

2. **Process**:
   - GitHub Actions checks out your repository
   - Copies the HTML file and markdown files to a deployment directory
   - Deploys everything to GitHub Pages

3. **Result**: Your mindmap is available at:
   ```
   https://[your-username].github.io/[repository-name]/
   ```

## Setup Instructions

### ⚠️ IMPORTANT: Enable GitHub Pages FIRST

**You must enable GitHub Pages before the workflow can run successfully!**

### Step 1: Enable GitHub Pages

1. Go to your repository on GitHub (e.g., `https://github.com/yourusername/system-design`)
2. Click on **Settings** (top menu bar of your repository)
3. In the left sidebar, click **Pages**
4. Under **Source**, you'll see a dropdown - select:
   - **Source**: `GitHub Actions` (NOT "Deploy from a branch")
   - This is crucial - it tells GitHub to use your workflow for deployment
5. **Save** or the page will auto-save

**Note**: If you don't see the "GitHub Actions" option:
- Make sure you have the workflow file (`.github/workflows/deploy-mindmap.yml`) in your repository
- Push the workflow file first, then come back to Settings → Pages
- You may need to refresh the page after pushing the workflow file

### Step 2: Verify Workflow File

The workflow file is located at:
```
.github/workflows/deploy-mindmap.yml
```

This file is already configured and ready to use.

### Step 3: Push Your Changes

Simply push your changes to the `main` branch:
```bash
git add .
git commit -m "Update mindmap"
git push origin main
```

The deployment will start automatically. You can monitor it in the **Actions** tab.

## File Structure

The deployment process:
- Copies `interactive-mind-map.html` → `index.html` (root of GitHub Pages)
- Copies `mind-maps/` directory → `mind-maps/` (preserves structure)
- The HTML file loads markdown files using relative paths like `mind-maps/interactive-mind-map.md`

## Accessing Your Mindmap

Once deployed, your mindmap will be available at:
- **URL**: `https://[username].github.io/[repo-name]/`
- **Example**: `https://yourusername.github.io/system-design/`

## Troubleshooting

### Error: "Get Pages site failed" or "Not Found"

**This error means GitHub Pages is not enabled yet!**

**Solution:**
1. Go to your repository → **Settings** → **Pages**
2. Under **Source**, select **GitHub Actions** (not "Deploy from a branch")
3. Save the settings
4. Go back to **Actions** tab and re-run the workflow (or push a new commit)

**Why this happens:**
- The workflow tries to configure Pages, but Pages must be enabled first in repository settings
- GitHub requires you to manually enable Pages before workflows can deploy to it

### Deployment Not Working

1. **Check Actions Tab**: Go to the **Actions** tab in your repository to see if the workflow ran and if there were any errors
2. **Check Permissions**: Ensure GitHub Pages is enabled in Settings → Pages with **Source: GitHub Actions**
3. **Check Workflow File**: Verify `.github/workflows/deploy-mindmap.yml` exists and is correct
4. **Verify Pages is Enabled**: In Settings → Pages, you should see "Your site is live at..." after enabling it

### Files Not Loading

1. **Check File Paths**: The HTML file uses relative paths. Make sure the `mind-maps/` directory structure is preserved
2. **Check Browser Console**: Open browser developer tools (F12) and check for any 404 errors
3. **Verify Deployment**: Check that files were actually deployed by visiting the raw file URLs

### Manual Deployment

If automatic deployment isn't working, you can manually trigger it:
1. Go to **Actions** tab
2. Select **Deploy Interactive Mindmap to GitHub Pages**
3. Click **Run workflow** → **Run workflow**

## Customization

### Change Deployment Branch

Edit `.github/workflows/deploy-mindmap.yml`:
```yaml
on:
  push:
    branches:
      - main  # Change this to your preferred branch
```

### Change Deployment Directory

If you want to deploy to a subdirectory (e.g., `/mindmap`), modify the workflow:
```yaml
- name: Prepare deployment files
  run: |
    mkdir -p _site/mindmap
    cp mind-map-interactive/interactive-mind-map.html _site/mindmap/index.html
    cp -r mind-map-interactive/mind-maps _site/mindmap/mind-maps
```

Then update the HTML file paths accordingly, or use a base tag in the HTML.

## Notes

- The first deployment may take a few minutes
- Subsequent deployments are faster (usually 1-2 minutes)
- GitHub Pages uses Jekyll by default, but since we're deploying static files, it works fine
- The HTML file uses CDN links for libraries (d3, markmap), so no build step is needed

