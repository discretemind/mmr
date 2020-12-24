# Merkle Mountain Range

## Introduction
The Merkle mountain range (MMR) had been invented by Peter Todd 
You can read about original implementation [here](https://github.com/opentimestamps/opentimestamps-server/blob/master/doc/merkle-mountain-range.md) and [here](https://github.com/mimblewimble/grin/blob/master/doc/mmr.md)

Current Implementation has another indexing, what makes navigation over the node much more easy and fast

![Mmr Structure](./doc/mmr-1.png)

 Blue Nodes represents a data objects linked by MMR structure.
 Green Modes represents supporting MMR nodes.
 
 Main advantage of MMR is that for N objects you have ~N supporting nodes. 
 So you can easily calculate data required fo storing data it's: `2*N` 
 
 Another feature: No need to have all data for adding new elements to MMR.
 
 ## Requirements
 
 Go 1.13+
 
 ## Install
 
 ```
 go get -u github.com/discretemind/mmr
 ```
 
 ## Example:
 
 ```
	m := mmr.Merkle(blake2b.New256)

	m.Add("test 1")
    m.Add("test 2")
    m.Add("test 3")
    m.Add("test 4")

    value3 := ""
	m.Get(3, &value3)
```

## License

Apache License v2