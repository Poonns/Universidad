""" El servicio de inform´atica de una universidad dispone de un fichero de texto con cierta informaci´on sobre los
accesos que los usuarios hacen a los ordenadores de las aulas de inform´atica durante cada d´ıa. Cada l´ınea del
fichero contiene, separados por un espacio en blanco, los siguientes datos: el n´umero IP que identifica el ordenador
al que se accede, la palabra succeeded o la palabra failed seg´un el acceso con login y password haya tenido ´exito
o no, y, en el caso de que el usuario haya accedido, el tiempo en minutos que utiliza el ordenador. A continuaci´on
se muestra el contenido de un posible fichero de ejemplo:
150.128.119.120 failed
150.128.204.131 failed
150.128.119.120 succeeded 20
150.128.119.120 succeeded 100
150.128.154.245 failed
150.128.154.245 succeeded 70
150.128.204.170 succeeded 60
150.128.154.245 succeeded 70
150.128.119.120 succeeded 20
150.128.154.245 failed """

""" a) (3 puntos) Una funci´on, crear_lista_accesos, que tenga como par´ametro el nombre de un fichero como los
descritos antes y devuelva como resultado una lista de objetos de la clase AccesosOrdenador con sus atributos
adecuadamente calculados. La lista contendr´a un ´unico objeto por cada ordenador (con IP diferente). La clase
est´a definida del siguiente modo: """
  
class AccesosOrdenador: 
    def __init__(self,ip):
        self.ip = ip
        self.intentos_login = 0
        self.login_con_exito = 0
        self.minutos_de_uso = 0

def crear_lista_accesos(nombre_fichero):
        lista_accesos = []

        file = open(nombre_fichero, 'r')

        for linea in file:
            partes = linea.strip().split()
            ip = partes[0]
            resultado = partes[1]
            if len(partes) > 2:
                minutos = int(partes[2])
            else:
                minutos = 0

            acceso_existente = None
            for acceso in lista_accesos:
                if acceso == ip:
                    acceso_existente = acceso
                    break
                
            if acceso_existente is None:
                nuevo_acceso = AccesosOrdenador(ip)
                nuevo_acceso.intentos_login = 1
                if resultado == 'succeeded':
                    nuevo_acceso.login_con_exito = 1
                    nuevo_acceso.minutos_de_uso = minutos
                    lista_accesos.append(nuevo_acceso)
                else:
                    acceso_existente.intentos_login += 1
                    if resultado == 'succeeded':
                        acceso_existente.login_con_exito += 1
                        acceso_existente.minutos_de_uso += minutos
        file.close()                
        return lista_accesos

""" b) (2 puntos) Una funci´on, ips_con_mayor_uso, que tenga como par´ametro una lista de accesos como la cons-
truida en el apartado anterior, que calcule cu´al es el mayor tiempo de uso de todos los ordenadores, y devuelva
una lista de cadenas con los n´umeros IP de los ordenadores que han tenido ese tiempo de uso. En el caso del
ejemplo, el tiempo m´aximo de uso es de 140 minutos, y como dos ordenadores han sido usados ese mismo
tiempo, la funci´on devuelve la lista [’150.128.119.120’, ’150.128.154.245’] """

def ips_con_mayor_uso(lista_accesos):
    lista_ips_mayor_uso = []
    mayor_tiempo_uso = 0
    
    for acceso in lista_accesos:
        if mayor_tiempo_uso < acceso.minutos_de_uso:
            mayor_tiempo_uso = acceso.minutos_de_uso
            lista_ips_mayor_uso = [acceso.ip]
        elif mayor_tiempo_uso == acceso.minutos_de_uso:
            lista_ips_mayor_uso.append(acceso.ip)

    return lista_ips_mayor_uso