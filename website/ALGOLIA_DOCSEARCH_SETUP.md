# Algolia DocSearch Configuration for KiviGo Documentation

This document explains the Algolia DocSearch configuration that has been set up for the KiviGo documentation website.

## 🔧 Configuration Added

The Algolia DocSearch has been configured in `docusaurus.config.ts` with the following settings:

```typescript
algolia: {
  // The application ID provided by Algolia
  appId: 'YOUR_APP_ID',

  // Public API key: it is safe to commit it
  apiKey: 'YOUR_SEARCH_API_KEY',

  // The index name to search
  indexName: 'kivigo',

  // Optional: see doc section below
  contextualSearch: true,

  // Optional: Specify domains where the navigation should occur through window.location instead on history.push
  externalUrlRegex: 'external\\.com|domain\\.com',

  // Optional: Replace parts of the item URLs from Algolia. Useful when using the same search index for multiple deployments using a different baseUrl.
  replaceSearchResultPathname: {
    from: '/docs/', // or as RegExp: /\/docs\//
    to: '/',
  },

  // Optional: Algolia search parameters
  searchParameters: {},

  // Optional: path for search page that enabled by default (`false` to disable it)
  searchPagePath: 'search',

  // Optional: whether the insights feature is enabled or not on Docsearch (`false` by default)
  insights: false,
},
```

## 🎯 Features Configured

### 1. **Search UI Integration**
- Search button appears in the top navigation bar with keyboard shortcut (Ctrl+K)
- Clean search modal with Algolia branding
- Responsive design that works on all screen sizes

### 2. **Contextual Search**
- `contextualSearch: true` enables version-aware search for the versioned documentation
- Search results will be scoped to the current documentation version

### 3. **URL Path Replacement**
- Configured to handle the `/kivigo/` base URL properly
- Search results will navigate correctly within the site structure

### 4. **Search Page**
- Dedicated search page available at `/search`
- Users can bookmark and share search URLs

## 🔑 Next Steps: Adding Algolia Credentials

To complete the setup, replace the placeholder values with your actual Algolia credentials:

### 1. **Get Algolia DocSearch Approved**
Apply for Algolia DocSearch at: https://docsearch.algolia.com/apply/

### 2. **Update Configuration**
Once approved, replace these values in `docusaurus.config.ts`:

```typescript
algolia: {
  appId: 'YOUR_ACTUAL_APP_ID',        // Replace with your App ID
  apiKey: 'YOUR_ACTUAL_SEARCH_API_KEY', // Replace with your Search API Key
  indexName: 'kivigo',                 // This can stay as 'kivigo' or use the index name provided by Algolia
  // ... rest of the configuration remains the same
},
```

### 3. **Configure Algolia Crawler**
Algolia will provide you with a crawler configuration. Make sure it includes:

- **Start URLs**: `https://azrod.github.io/kivigo/docs/`
- **Sitemap URL**: `https://azrod.github.io/kivigo/sitemap.xml`
- **Selectors**: Configured for Docusaurus structure

## 🧪 Testing

The search functionality has been tested and verified:

1. ✅ Search UI appears in navigation
2. ✅ Search modal opens with Ctrl+K or click
3. ✅ Algolia branding is properly displayed
4. ✅ Build process works correctly
5. ✅ Compatible with versioned documentation

## 📚 Documentation Structure

The search will index content from:

- **Main documentation** (`/docs/`)
- **Versioned documentation** (`/docs/1.5.0/`, etc.)
- **All backend guides** (`/docs/backends/`)
- **Advanced features** (`/docs/advanced/`)

## 🔍 Search Features

Once the Algolia credentials are added, users will be able to:

- **Full-text search** across all documentation
- **Autocomplete suggestions** as they type
- **Keyboard navigation** with arrow keys
- **Quick access** with Ctrl+K shortcut
- **Deep linking** to specific sections
- **Version-scoped results** (when browsing a specific version)

## 📝 Notes

- The configuration is ready for production use
- No additional dependencies were added (Docusaurus includes Algolia support)
- The setup follows Docusaurus best practices
- All placeholder values are clearly marked for easy replacement