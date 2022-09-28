# Pipeline Functions

## Tag Resolve

Implementation can be found on [/cmd/clade/cmd/internal/load.go](/cmd/clade/cmd/internal/load.go).

### localTags

Returns the tags created in the current tree. If the image is the root node (if it is a remote image), it will return an empty array.

### remoteTags

Returns the tags fetched from the registry.

## Filter by Regex

Implementation can be found on [/plf/regex.go](/plf/regex.go).

### regex

Returns matched strings by regex.
It is allowed to use subexpressions defined by [regexp](https://pkg.go.dev/regexp).

```
regex foo(.*)ba(?P<baz>.*)r fobar fooaaabar foobaaar baz
```
The above returns `fooaaabar` and `foobaaar`.
The captured string can be accessed by index in the `_` field or by the same field name.
For example, `{{ index ._ 1 }}` would be `aaa` for `fooaaabar`, and `{{ .baz }}` would be `aa` for `foobaaar`
First index of `_` holds original string; `{{ index ._ 0 }}` for `fooaaabar` would be `fooaaabar`.


## Filter by *Semantic Versioning*

Implementation can be found on [/plf/regex.go](/plf/semver.go).

## toSemver

Returns *Semantic Versioning*-like strings.

```
toSemver 12.1.2 12.1-alpine3.6 noetic
```
The above returns `12.1.2` and `12.1-alpine3.6`.
Note that `12.1-alpine3.6` has no *patch* version and has invalid *pre-release* version but it is evaluated as valid.
Each part of the version can be accessed by its name.
For example, `{{ printf "%d.%d.%d" .Major .Minor .Patch }}` for `12.1-alpine3.6` would be `12.1.0`.
Access to *pre-release* or *build metadata* is not defined (unstable).

## semverLatest

Returns a latest version.

```
toSemver 1.0 1.0.2 2.0 1.3.1 | semverLatest
```
The above returns `2.0`.
Only the `sv.Version` type returned by a function such as `toSemver` is accepted as an argument.

## semverMajorN

Returns last *N* major versions.

```
toSemver 1 2 3 2.0 | semverMajorN 2
```
The above returns `2`, `3`, and `2.0`.

## semverMinorN

Returns last *N* minor versions.

```
toSemver 1.0 1.5 1.2 3.2 2.6 2.9 2.13 | semverMinorN 2
```
The above returns `1.5`, `1.2`, `3.2`, `2.9`, and `2.13`.

## semverPathN

Returns last *N* patch versions.
```
toSemver 1.0 1.0.4 1.0.5 2.1.5 2.2.6 3.1.6 3.1.4 3.1.42 | semverPatchN 2
```
The above returns `1.0.4`, `1.0.5`, `2.1.5`, `2.2.6`, `3.1.6` and `3.1.42`.

## semverN

Takes 3 numbers such that *X*, *Y*, *Z*, and returns last *X* major versions, *Y* minor versions, *Z* patch versions.

```
toSemver 1.0 2.5 3.2.6 3.14.2 3.6.9 | semverN 2 2 1
```
The  above returns `3.14.2`.
This function effectively equals to:
```
semverMajorN X | semverMinorN Y | semverPatch Z
```
