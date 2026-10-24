# SwiftData

Read before writing or reviewing `@Model` types, `@Query`, `#Predicate`, `ModelContainer` / `ModelContext` code, or anything that syncs SwiftData with CloudKit.

Contents

1. Ground rules
2. Defining models
3. Relationships and delete rules
4. Saving and identifiers
5. `@Query` and fetching
6. Predicates
7. Indexes and uniqueness
8. Model inheritance
9. CloudKit
10. Concurrency and `ModelActor`
11. Previews and in-memory containers

## 1. Ground rules

- Check the deployment target first: indexes need iOS 18+, model subclassing iOS 26+.
- Use SwiftData throughout. Do not reach for Core Data unless SwiftData cannot do the job at all.
- Do not add third-party persistence frameworks without asking.
- Read whether the container is configured for CloudKit before reviewing models; it changes several rules (section 9).

## 2. Defining models

- Never name a property `description` in a `@Model` class; it is disallowed.
- Do not add `didSet` / `willSet` to `@Model` properties; they are silently ignored. Put the side effect in a method that sets the property.
- Stored enums must be `Codable`. Enums with associated values work.
- `@Attribute(.externalStorage)` is a hint, only for `Data`; SwiftData may ignore it.
- `@Transient` properties are not persisted, need a default, and reset to it on every fetch. Prefer a computed property when the value derives from stored ones; use `@Transient` only for expensive values.
- Define a `VersionedSchema` and `SchemaMigrationPlan` from the first release, even if every migration is lightweight.

## 3. Relationships and delete rules

- Write `@Relationship` on one side only. Putting it on both sides creates a circular reference.
- Always name the inverse explicitly; SwiftData's inference often gets it wrong.
- Always give an explicit delete rule. The default `.nullify` leaves orphans and crashes when the other side is non-optional. Usually `.cascade` for owned children.

```swift
// Before
var sights: [Sight]

// After
@Relationship(deleteRule: .cascade, inverse: \Sight.destination)
var sights: [Sight]
```

## 4. Saving and identifiers

- Autosave timing is unpredictable. Call `try modelContext.save()` explicitly wherever correctness depends on the data being persisted.
- Call `save()` directly; do not guard it with `hasChanges`.
- A `PersistentIdentifier` is temporary until the first save (temporary IDs start with a lowercase `t`) and changes on save. Save before storing, sending or comparing an ID.

## 5. `@Query` and fetching

- Use `@Query` only inside SwiftUI views. Anywhere else, fetch from a `ModelContext`:

```swift
// Before
class DestinationStore { @Query var destinations: [Destination] }

// After
class DestinationStore {
    let modelContext: ModelContext
    func destinations() throws -> [Destination] {
        try modelContext.fetch(FetchDescriptor<Destination>())
    }
}
```

- For a query that depends on view input, build it in the view's `init`: `_items = Query(filter: predicate, sort: \.name)`.
- For a count only, use `modelContext.fetchCount(_:)`. It does not live-update; something else (an `@Query`, a refresh) must trigger re-reading.
- On a `FetchDescriptor`:
  - set `relationshipKeyPathsForPrefetching` for relationships you know you will read;
  - set `propertiesToFetch` to load only the properties you use;
  - set `fetchLimit` / `fetchOffset` for paging rather than fetching all and slicing.

## 6. Predicates

`#Predicate` supports a subset of Swift. Some unsupported code fails to compile; some compiles and then fails or crashes at runtime. Write only what is listed as safe here.

Safe:

- Text search: `localizedStandardContains(_:)`, never `lowercased().contains(…)`.
- Prefix match: `starts(with:)`. `hasPrefix` / `hasSuffix` are unsupported.
- Empty check: `!$0.items.isEmpty`.
- Type filtering with subclasses: `is` and `as?` (section 8).
- External values: copy them into a local `let` before the predicate and reference the local.

Does not compile: `hasSuffix`, `lowercased()`, `map`, `reduce`, `count(where:)`, `first`, custom operators.

Compiles, then fails or crashes at runtime. Never write:

- `$0.items.isEmpty == false`; write `!$0.items.isEmpty`.
- Filters on computed properties, `@Transient` properties, or the inside of `Codable` struct values. Predicates can only see data actually stored as model attributes. If you must filter on a value stored inside a `Codable` type, store that value as its own plain attribute too.
- Regex literals (`$0.name.contains(/Titanic/)`).

## 7. Indexes and uniqueness

- `#Index` (iOS 18+) speeds reads and costs a little on every write. Skip it for write-heavy, rarely-read data such as logs.
- Index single properties and, separately, groups that are queried together:

```swift
@Model class Article {
    #Index<Article>([\.type], [\.type, \.author])
    var type: String
    var author: String
}
```

- Write one `#Unique` per model, inside the class. Pass several key-path arrays to it for several constraints: `#Unique<User>([\.email], [\.username])`.
- Never use `#Unique` or `@Attribute(.unique)` with CloudKit (section 9).

## 8. Model inheritance

Needs iOS 26+. Use only when it pays off; a protocol is usually simpler, and deep hierarchies make migrations harder.

- Mark both parent and child `@Model`.
- Mark every subclass `@available(iOS 26, *)` even when the deployment target is already iOS 26. Code that uses the subclass needs no extra annotation when the deployment target covers it.
- List the parent and every subclass in the container's schema; SwiftData does not discover subclasses.
- A relationship typed as the parent may hold any subclass. If only one subclass belongs there, type the relationship as that subclass.
- `@Query var tutorials: [Tutorial]` returns only that subclass; `@Query var articles: [Article]` returns the parent and all subclasses.
- Filter to some subclasses with `is`; filter on subclass properties with `as?` inside the predicate. Results are typed as the parent, so cast to reach subclass members.

```swift
@Query(filter: #Predicate<Article> { article in
    if let tutorial = article as? Tutorial { tutorial.difficulty < 3 }
    else { article is News }
}) private var frontPage: [Article]
```

## 9. CloudKit

Applies only when the container syncs through CloudKit.

- No `@Attribute(.unique)` or `#Unique`. They are unsupported, and using them breaks the local store too.
- Every property needs a default value or must be optional.
- Every relationship must be optional.
- Indexes and subclasses are fine on OS versions that support them.
- Sync is eventually consistent. Code must work when related records or recent changes have not arrived yet: handle missing relationships, and never assume a just-written record exists on another device.

## 10. Concurrency and `ModelActor`

- Never pass a `ModelContext` or a `@Model` instance across actors. Neither is `Sendable`.
- `ModelContainer` and `PersistentIdentifier` are `Sendable`. Send the identifier, then re-fetch in the destination context with `context.model(for:)` or a `FetchDescriptor` filtered on `persistentModelID`. Save first so the ID is permanent (section 4).
- The view's `@Environment(\.modelContext)` is a main-actor context. Use it only from main-actor code.
- For background imports and bulk work, use a `@ModelActor` actor built from the shared container; it owns its own `modelContext`:

```swift
@ModelActor
actor Importer {
    func importItems(_ dtos: [ItemDTO]) throws {
        for dto in dtos { modelContext.insert(Item(dto)) }
        try modelContext.save()
    }
}
```

- Create a `@ModelActor` from a non-main context (for example inside a `@concurrent` function). One initialised on the main actor can end up running its work on the main thread.
- Pass `Sendable` DTOs or identifiers into and out of the actor, never model instances. Return identifiers and let the main context re-fetch.
- Save in the background context so the main context and `@Query` views pick up the changes.
- For the wider rules on isolation and `Sendable`, read `concurrency.md`.

## 11. Previews and in-memory containers

- Give previews and throwaway contexts an in-memory store so they never touch the real database: `.modelContainer(for: Item.self, inMemory: true)`, or build `ModelContainer(for: Item.self, configurations: ModelConfiguration(isStoredInMemoryOnly: true))` and insert sample data into its `mainContext`.
- Share one sample-data container across previews (a static factory, or a `PreviewModifier` on iOS 18+) instead of rebuilding it in every `#Preview`.
- List every model type the preview touches, including subclasses and related models, in the container's schema.
