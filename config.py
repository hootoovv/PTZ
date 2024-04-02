import yaml

def loadConfig():
  with open('config.yaml', 'r') as ymlfile:
    cfg = yaml.load(ymlfile, Loader=yaml.FullLoader)
    return cfg['ip'], cfg['port'], cfg['username'], cfg['password']