# Pipeline Functions

## Default Functions from `lesomnus/pl`

Implementation can be found on [lesomnus/pl/funcs](https://github.com/lesomnus/pl/tree/main/funcs).


## Tag Resolve

Implementation can be found on [/cmd/clade/cmd/internal/load.go](/cmd/clade/cmd/internal/load.go).

### tags

Returns the tags fetched from the registry if the current image depend on the remote image;
if the current image depend on the generated image, it returns the tags generated in the current tree.


## Filter by *Semantic Versioning*

Implementation can be found on [/plf/regex.go](/plf/semver.go).

### semver

Returns *Semantic Versioning*-like strings.

```
( semver "12.1.2" "12.1-alpine3.6" "noetic" )
```
The above returns `12.1.2` and `12.1-alpine3.6`.
Note that `12.1-alpine3.6` has no *patch* version and has invalid *pre-release* version but it is evaluated as valid.
Each part of the version can be accessed by its name.
For example, `( printf "%d.%d.%d" $.Major $.Minor $.Patch )` for `12.1-alpine3.6` would be `12.1.0`.
Access to *pre-release* or *build* version is not defined (unstable).

### semverFinalized

Returns *Semantic Versioning*-like strings that does not contains *pre-release* or *build* versions.

```
( semver "12.1.2" "12.1-alpine3.6" "noetic" )
```
The above returns `12.1.2`.

### semverLatest

Returns a latest version.

```
( semverLatest "1.0" "1.0.2" "2.0" "1.3.1" )
```
The above returns `2.0`.
Only the `sv.Version` type returned by a function such as `toSemver` is accepted as an argument.

### semverMajorN

Returns last *N* major versions.

```
( semverMajorN 2 "1" "2" "3" "2.0" )
```
The above returns `2`, `3`, and `2.0`.

### semverMinorN

Returns last *N* minor versions.

```
( semverMinorN 2 "1.0" "1.5" "1.2" "3.2" "2.6" "2.9" "2.13" )
```
The above returns `1.5`, `1.2`, `3.2`, `2.9`, and `2.13`.

### semverPathN

Returns last *N* patch versions.
```
( semverPatchN 2 "1.0" "1.0.4" "1.0.5" "2.1.5" "2.2.6" "3.1.6" "3.1.4" "3.1.42" )
```
The above returns `1.0.4`, `1.0.5`, `2.1.5`, `2.2.6`, `3.1.6` and `3.1.42`.

### semverN

Takes 3 numbers such that *X*, *Y*, *Z*, and returns last *X* major versions, *Y* minor versions, *Z* patch versions.

```
( semverN 2 2 1 "1.0" "2.5" "3.2.6" "3.14.2" "3.6.9" )
```
The  above returns `3.14.2`.
This function effectively equals to:
```
( semverMajorN X | semverMinorN Y | semverPatch Z )
```
