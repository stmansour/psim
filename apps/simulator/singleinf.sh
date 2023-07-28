#!/bin/bash

influencers=(
    "CCInfluencer"
    "DRInfluencer"
    "URInfluencer"
    "MSInfluencer"
)

# Backup the original config.json5 file
cp config.json5 config.bak

#---------------------------------------------------------
# Iterate through the influencers, run a simulation with
# it as the one-and-only influencer.  Save the results.
#---------------------------------------------------------
for influencer in "${influencers[@]}"; do
    echo "Simulation with only ${influencer}"

    #-----------------------------------------------------------------
    # Generate config.json5 with ${influencer} as the only influencer
    #-----------------------------------------------------------------
    awk -v influencer="$influencer" '
        /"InfluencerSubclasses": \[/ { print; found = 1; next }
        found && /    \],/ { print "        \"" influencer "\""; print; found = 0; next }
        !found { print }
    ' config.bak > config.json5.tmp && mv config.json5.tmp config.json5

    ./simulator
    mv SimStats.csv "${influencer}.csv"
done

# Cleanup
# mv config.bak config.json5
