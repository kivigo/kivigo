# Contributing to KiviGo

Thank you for your interest in contributing to KiviGo!  
Your help is welcome, whether it's for fixing bugs, adding features, improving documentation, or sharing feedback.

## How to Contribute

### 1. Fork & Clone

- Fork the repository on GitHub.
- Clone your fork locally:

  ```sh
  git clone https://github.com/your-username/kivigo.git
  cd kivigo
  ```

### 2. Create a Branch

- Create a new branch for your change:

  ```sh
  git checkout -b my-feature
  ```

### 3. Make Your Changes

- Write clear, concise code.
- Add or update tests if needed.
- Follow the existing code style (see `.golangci.yml` for linting rules).
- Document your code and update the `README.md` or package docs if necessary.

### 4. Test

- Run all tests before submitting:

  ```sh
  go test ./...
  ```

### 5. Lint

- Run the linter to check for style issues:

  ```sh
  golangci-lint run
  ```

### 6. Commit & Push

- Write clear commit messages.
- Push your branch to your fork:

  ```sh
  git push origin my-feature
  ```

### 7. Open a Pull Request

- Go to the GitHub page of your fork and open a Pull Request (PR) against the `main` branch.
- Describe your changes and reference any related issues.

## 🚀 Creating Releases

For maintainers: KiviGo uses an automated release workflow that creates tags for both the main library and all backend modules.

### Using the Release Workflow

1. Go to the **Actions** tab in the GitHub repository
2. Select the **Release** workflow
3. Click **Run workflow**
4. Configure the release options:
   - **Version**: Enter the version (e.g., `v1.5.0`) following semantic versioning
   - **Release type**: Choose what to release:
     - `core+backends` (default): Creates both main project and backend tags
     - `core`: Creates only the main project tag
     - `backends`: Creates only backend module tags
   - **Mark as latest**: Check to mark this release as the latest (default: unchecked)
   - **Publish documentation**: Check to publish documentation (default: checked)

The workflow will automatically:

- Create tags based on your selection (main project and/or backend-specific tags)
- Deploy documentation (for core releases, if enabled)

#### Release Type Examples

**Core + Backends (default)**:

- Creates main project tag: `v1.5.0`
- Creates backend tags: `backend/consul/v1.5.0`, `backend/redis/v1.5.0`, etc.
- Deploys documentation (if enabled)

**Core only**:

- Creates only main project tag: `v1.5.0`
- Deploys documentation (if enabled)

**Backends only**:

- Creates only backend tags: `backend/consul/v1.5.0`, `backend/redis/v1.5.0`, etc.
- No documentation deployment

### Manual Tagging (Alternative)

If needed, you can create tags manually:

```sh
# Main project
git tag v1.5.0

# Backend modules (repeat for each backend)
git tag backend/consul/v1.5.0
git tag backend/redis/v1.5.0
# ... etc

# Push all tags
git push origin --tags
```

### Documentation Deployment

The documentation deployment is available as an independent workflow that can be run separately:

1. Go to the **Actions** tab in the GitHub repository
2. Select the **Deploy Documentation** workflow
3. Click **Run workflow**
4. Enter the version for which to deploy documentation (e.g., `v1.5.0`)

This is useful when you need to:

- Deploy documentation without creating a release
- Re-deploy documentation for an existing version
- Update documentation after fixing content issues

## Code of Conduct

Please be respectful and constructive in all interactions.  
See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

## Need Help?

- Open an [issue](https://github.com/kivigo/kivigo/issues) for questions, bugs, or feature requests.
- For major changes, please open an issue first to discuss what you would like to change.

Thank you for helping to make KiviGo better!
