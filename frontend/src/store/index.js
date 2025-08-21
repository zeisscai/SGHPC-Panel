import { createStore } from 'vuex'

export default createStore({
  state: {
    managementNode: {},
    computeNodes: [],
    slurmJobs: []
  },
  mutations: {
    SET_MANAGEMENT_NODE(state, node) {
      state.managementNode = node
    },
    SET_COMPUTE_NODES(state, nodes) {
      state.computeNodes = nodes
    },
    SET_SLURM_JOBS(state, jobs) {
      state.slurmJobs = jobs
    }
  },
  actions: {
    updateManagementNode({ commit }, node) {
      commit('SET_MANAGEMENT_NODE', node)
    },
    updateComputeNodes({ commit }, nodes) {
      commit('SET_COMPUTE_NODES', nodes)
    },
    updateSlurmJobs({ commit }, jobs) {
      commit('SET_SLURM_JOBS', jobs)
    }
  },
  getters: {
    getManagementNode: (state) => state.managementNode,
    getComputeNodes: (state) => state.computeNodes,
    getSlurmJobs: (state) => state.slurmJobs
  }
})