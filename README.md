# distributed-sys-project
- First feature we'll be tackling is the the feature of high availability for writes 
- We would need to be able to spawn multiple nodes and have multiple properties for each node (within the metadata)
- We done the above already, now we need to have a ring server to make sure that each node knows who their successor and predecessor is
- Currently we still have a single point of failure as the 'node server' is the one handling the hashing and the delegation of resources.
- frontend is nice