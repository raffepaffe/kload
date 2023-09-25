// Kload shows cpu and memory load on nodes and pods in your kubernetes cluster. Kload uses your kubectl setup
// to show information about your cluster.
//
// Basic usage:
//
// Show load in a namespace:
// kload -pod -ns [your-namespace]
//
// Show load in a namespace for a pod:
// kload -pod -ns [your-namespace] [your-pod]
//
// Show load for nodes:
// kload -node [node-name]
//
// Find more information at https://github.com/raffepaffe/kload/blob/master/README.md
package main
