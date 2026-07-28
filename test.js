const edgesData = [{"consumer":"mcp-org/frontend","provider":"mcp-org/backend","status":"active","consumer_metadata":{},"provider_metadata":{}},{"consumer":"mcp-org/frontend","provider":"mcp-org/backend","status":"active","consumer_metadata":{},"provider_metadata":{}}];

let rawNodes = [];
let rawEdges = [];

const processGraphData = (edgesData) => {
    let newNodesMap = new Map();
    let newEdges = [];

    const addTeamNode = (teamName) => {
        const teamId = `team-${teamName}`;
        if (!newNodesMap.has(teamId)) {
            newNodesMap.set(teamId, {
                id: teamId,
                type: 'teamGroup',
                position: { x: 0, y: 0 },
                data: { label: teamName }
            });
        }
    };

    const addNode = (id, status, type, metadata = {}) => {
        let teamName = metadata?.team || null;
        if (teamName) addTeamNode(teamName);

        if (!newNodesMap.has(id)) {
            const existingNode = rawNodes.find(n => n.id === id);
            const volatilityScore = existingNode?.data?.volatilityScore !== undefined
                ? existingNode.data.volatilityScore
                : Math.floor(Math.random() * 101);

            const node = {
                id,
                type: 'service',
                position: { x: 0, y: 0 },
                data: { label: id, status, type, metadata, volatilityScore }
            };
            if (teamName) {
                node.parentId = `team-${teamName}`;
                node.extent = 'parent';
            }
            newNodesMap.set(id, node);
        } else {
            const existing = newNodesMap.get(id);
            if (existing) {
                if (status === 'BREAKING') existing.data.status = 'BREAKING';
                if (metadata && Object.keys(metadata).length > 0) {
                    existing.data.metadata = metadata;
                }
                if (teamName && !existing.parentId) {
                    existing.parentId = `team-${teamName}`;
                    existing.extent = 'parent';
                }
            }
        }
    };

    edgesData.forEach((edge) => {
        addNode(edge.provider, edge.status, 'provider', edge.provider_metadata);
        addNode(edge.consumer, 'SAFE', 'consumer', edge.consumer_metadata);

        newEdges.push({
            id: `e-${edge.provider}-${edge.consumer}`,
            source: edge.provider,
            target: edge.consumer,
            type: edgesData.length < 150 ? 'interactive' : 'straight',
            animated: true,
            style: `stroke: ${edge.status === 'BREAKING' ? '#EF4444' : '#64748b'}; stroke-width: 2px;`
        });
    });

    let nextNodes = Array.from(newNodesMap.values());

    const globalBreakingImpacts = new Set();
    for (const n of nextNodes) {
        if (String(n.data.status).toLowerCase() === 'breaking') globalBreakingImpacts.add(n.id);
    }

    let changed = true;
    while (changed) {
        changed = false;
        for (const edge of newEdges) {
            if (globalBreakingImpacts.has(edge.source) || edge.style?.includes('#EF4444')) {
                if (!globalBreakingImpacts.has(edge.target)) {
                    globalBreakingImpacts.add(edge.target);
                    changed = true;
                }
            }
        }
    }

    rawNodes = nextNodes.map(n => {
        if (globalBreakingImpacts.has(n.id) && String(n.data.status).toLowerCase() !== 'breaking') {
            return {
                ...n,
                data: { ...n.data, status: 'BREAKING' }
            };
        }
        return n;
    });

    rawEdges = newEdges;
    console.log('processGraphData finished. rawNodes length:', rawNodes.length);
};

processGraphData(edgesData);
console.log(rawNodes);
