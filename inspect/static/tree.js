import * as vis from "https://unpkg.com/vis-network/standalone/esm/vis-network.min.mjs";

const visOptions = {
    physics: {
        enabled: true,
        barnesHut: { gravitationalConstant: -2000, avoidOverlap: 0.2 },
        solver: "hierarchicalRepulsion",
        hierarchicalRepulsion: {
            nodeDistance: 140,
        },
        stabilization: {
            enabled: true,
            iterations: 200,
            updateInterval: 25
        },
    },
    layout: {
        randomSeed: 1,
    },
    edges: {
        smooth: {
            type: 'curvedCW', // Options: 'curvedCW', 'curvedCCW', 'dynamic', etc.
            roundness: 0.2 // Adjust curvature
        },
        arrows: {
            to: { enabled: true, scaleFactor: 1 }
        }
    }
};

(async () => {
    const treeData = await fetch("/pal/tree.json").then(res => res.json());

    // treeData.nodes.sort((a, b) => a.inDegree - b.inDegree);
    treeData.nodes.sort((a, b) => a.label.localeCompare(b.label));

    treeData.nodes.forEach(node => {
        if (node.runner) {
            node.shape = "box";
        }
    });

    // Create a network
    const container = document.getElementById("tree");
    const data = {
        nodes: new vis.DataSet(treeData.nodes),
        edges: new vis.DataSet(treeData.edges),
    };


    // Initialize the network
    const network = new vis.Network(container, data, visOptions);
})();