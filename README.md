# MeshParser 
#### A RobloxMesh parser written in pure go

# WARNING READING MESH v1 IS DISABLED
Mesh version 1 caused issues when porting forward and i honestly cba fixing it

If you can fix it and push a fix that will be appreciated. The issue is with faces
I decided to disable it since there isnt much benifit to converting to modern versions since its such an old and supported format.

Any client past 2010 can parse Mesh V2 and it is a smaller and faster format

## Suported Versions

- Mesh V2 (NoRgba & Rgba)
- Mesh V3
- Mesh V4 & V4.1


## Future plans

- Mesh V5  (If it becomes an issue not being able to parse them)
- Mesh V6 (If i can find specifications + an actual mesh)
- Fixing Mesh V1

## Usage

#### Requires a version of go with go.mod support

1. Add to project 
```bash
go get github.com/MojaveMF/MeshParser
```

2. Import 
```go
import mesh "github.com/MojaveMF/MeshParser"
```

## Examples

### Converting a mesh

```go
import mesh "github.com/MojaveMF/MeshParser"

mesh,err := mesh.DecodeMesh(stream)
if err != nil {
    /* Handle err */
}

if err := mesh.ExportV2().Write(output); err != nil {
    /* Handle err */
}

```

### Converting a mesh only if needed

```go
import mesh "github.com/MojaveMF/MeshParser"
decodeLayer := mesh.MeshDecodeLayer(mesh.MeshVersion4, mesh.MeshVersion2)

/* Input and Output streams */
if err := decodeLayer(input,output); err != nil {
    /* Handle err */
}
```

## Why streams?
This is designed to be used on a webserver and often times the data is streamed back and forth from client to client. I do this since i believe it to be more efficent than large slices.