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
4. Enter the version (e.g., `v1.5.0`) following semantic versioning
5. Choose whether to create a GitHub release (recommended: `true`)

The workflow will automatically:
- Create a main project tag (e.g., `v1.5.0`)
- Create backend-specific tags (e.g., `backend/consul/v1.5.0`, `backend/redis/v1.5.0`)
- Generate a GitHub release with usage instructions

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

## Code of Conduct

Please be respectful and constructive in all interactions.  
See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

## Need Help?

- Open an [issue](https://github.com/azrod/kivigo/issues) for questions, bugs, or feature requests.
- For major changes, please open an issue first to discuss what you would like to change.

Thank you for helping to make KiviGo better!
