"""
Generates a Paxos configuration given the input parameters.
Prints the generated configuration to standard output
"""

import sys

REPLICA_BASE_ID = 1
REPLICA_BASE_PORT = 5000
CLIENT_BASE_ID = 100
CLIENT_BASE_PORT = 8000
CONTROLLER_ID = 1000
CONTROLLER_PORT = 9999


class Agent(object):

    def __init__(self, id, kind, port):
        self.id = id
        self.kind = kind
        self.box = f"127.0.0.1:{port}"
        self.attrs = {}  
        self.routes = []  # list of (int, int), i.e. (phydest, port)

    def gen_attr_lines(self):
        res = []
        for k in self.attrs.keys():
            if type(self.attrs[k]) == str:
                # if it's a string attr, wrap it in quotes
                res.append(f"\"{k}\" : \"{self.attrs[k]}\"")
            else:
                res.append(f"\"{k}\" : {str(self.attrs[k])}")
        return res

    def gen_route_lines(self):
        res = []
        for route in self.routes:
            dest = route[0]
            port = route[1]
            res.append(f"\"{dest}\" : " + "{" + f" \"{dest}\" : {port} " + "}")
        return res

    def __str__(self):
        return "\n".join(
            [f"\t\"{self.id}\" : " + " {", 
            f"\t\t\"type\" : \"{self.kind}\",",
            f"\t\t\"box\" : \"{self.box}\",",
            f"\t\t\"attrs\" : " + "{",
            f"\t\t\t" + f",\n\t\t\t".join(self.gen_attr_lines()),
            "\t\t},",
            f"\t\t\"routes\" : " + "{",
            f"\t\t\t" + f",\n\t\t\t".join(self.gen_route_lines()),
            "\t\t}",
            "\t}"
            ]
        )


def print_agents(agents):
    print("{")
    print(",\n".join([str(ag) for ag in agents]))
    print("}")


def generate(f, num_clients, client_mode):
    """
    Generates and prints a Paxos configuration to stdout
    :param f: Number of replica failures the paxos configuration tolerates
    :num_cliends: Number of client agents desired in the configuration
    :client_mode: Is this client in 'manual' or 'script' mode
    """
    assert f > 0 
    assert num_clients > 0
    assert client_mode == 'script' or client_mode == 'manual'

    # Generate agent objects
    replicas = [REPLICA_BASE_ID + i for i in range(2*f + 1)]
    clients = [CLIENT_BASE_ID + i for i in range(num_clients)]
    agents = []
    for rep in replicas:
        agents.append(Agent(rep, 'paxos_replica', REPLICA_BASE_PORT + rep))
    for clt in clients:
        agents.append(Agent(clt, 'paxos_client', CLIENT_BASE_PORT + clt))
    
    # Populate agent attributes
    for agent in agents:
        if agent.kind == 'paxos_replica':
            agent.attrs["myid"] = agent.id
            agent.attrs["replicas"] = replicas
            agent.attrs["clients"] = clients
            agent.attrs["output"] = f"tmp/replica_{agent.id}.output"
        else:  # agent.kind == 'client
            agent.attrs["myid"] = agent.id
            agent.attrs["replicas"] = replicas
            agent.attrs["mode"] = client_mode

    # Populate agent routing tables
    for agent in agents:
        if agent.kind == 'paxos_replica':
            # Connect to every replica and every client
            for rep in replicas:
                agent.routes.append((rep, 1))  # replicas listen for replicas on port 1
            for clt in clients:
                agent.routes.append((clt, 1))  # clients listen for replicas on port 1
        else:  # agent.kind == 'client
            for rep in replicas:
                agent.routes.append((rep, 2))  # replicas listen for client on port 2
    
    # Add controller agent
    controller = Agent(999, "paxos_controller", 9999)
    controller.attrs["replicas"] = replicas
    controller.attrs["clients"] = clients
    for rep in replicas:
        controller.routes.append((rep, 9))  # all agents listen for controller on port 9
    for clt in clients:
        controller.routes.append((clt, 9))
    agents.append(controller)
    print_agents(agents)


if __name__ == '__main__':
    f = int(sys.argv[1])
    num_clients = int(sys.argv[2])
    client_mode = sys.argv[3] 
    generate(f, num_clients, client_mode)
