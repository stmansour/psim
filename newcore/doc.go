// Package newcore is the module that implements the core of the simulator.
// Influencers are objects that make predictions based on a particular metric.
// All influencers implement the Infleuencer interface. Currently all
// influencers are a implemented within a single subclass of Influencer called
// LSMInfluencer.  A separate LSMInfluencer pseudo-subclass is created for each
// metric defined in the MISubclass table.
// Factory is used to handle the majority of the genetic-related operations. These
// include the creation of new populations, creating Investors and Influencers from
// DNA strings, and mutating investors and influencers.
package newcore
