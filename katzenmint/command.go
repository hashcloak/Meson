package katzenmint

type Command uint8

var (
	PublishMixDescriptor Command = 1
	AddConsensusDocument Command = 2 // Deprecated
	AddNewAuthority      Command = 3
	GetConsensus         Command = 4
	GetEpoch             Command = 5
)
