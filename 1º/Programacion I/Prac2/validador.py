#!/usr/bin/env python
'''
Created on 16/07/2014

@version 1.3.8 (2020-11-15)
@author: David Llorens (dllorens@uji.es)
         Federico Prat (fprat@uji.es)
         Juan Miguel Vilar (jvilar@uji.es)

@maintainer: Juan Miguel Vilar (jvilar@uji.es)

Comprobador simple de programas
-------------------------------
Si se ejecuta en versiones con sigalrm y TimeoutError utilizará timeouts al lanzar las ejecuciones.
Esto permite que programas mal construidos (p.e. con bucles infinitos) no bloqueen al
propio comprobador.
'''

import ast
import codecs
import copy
import glob
import inspect
from io import StringIO
import os
import signal
from sys import version_info, executable
import sys
import traceback
import subprocess
import zipfile


def error(m):
    print("ERROR:", m)
    sys.exit()

def clean(s):
    return " ".join(s.strip().lower().split())

def my_deepcopy(v):
    """Perform a deep copy if v is not a module, else return it
    untouched"""
    if type(v) != type(sys):
        return copy.deepcopy(v)
    else:
        return v

def comparaSalida(user, out):
    lineasUser = [ clean(l) for l in user.split("\n") if l.strip()!="" ]
    lineasOut  = [ clean(l) for l in out.split("\n") if l.strip()!="" ]

    longitud = min(len(lineasUser), len(lineasOut))
    for i in range(longitud):
        if lineasUser[i] != lineasOut[i]:
            return True, lineasUser[i], lineasOut[i]

    if len(lineasUser) == len(lineasOut):
        return False, None, None
    elif len(lineasUser) > longitud:
        return True, "\n".join(lineasUser[longitud:]), None
    else:
        return True, None, "\n".join(lineasOut[longitud:])

def matriz(datos):
    lineas=datos.split('\n')
    mat=[]
    for l in lineas:
        if l=="": break
        mat.append([int(n) for n in l.split()])
    return mat

def leeLíneas(nombre):
    with open(nombre) as f:
        return f.readlines()

def sing_pl(n, sing, pl):
    if n == 1:
        return sing
    else:
        return pl.format(n)

def compararFicheros(expected, user):
    try:
        userLines = leeLíneas(user)
    except:
        return "Error leyendo el fichero {}".format(user)
    try:
        expectedLines = leeLíneas(expected)
    except:
        return "Error leyendo el fichero de pruebas {}, comprueba que existe.".format(user)
    luser = len(userLines)
    lexpected = len(expectedLines)
    for i in range(min(luser, lexpected)):
        if userLines[i] != expectedLines[i]:
            return ("La línea {} del fichero {} no coincide con lo esperado."
                    " Es:\n{}\ny debería ser:\n{}").format(i+1, user, userLines[i], expectedLines[i])
    if luser < lexpected:
        return "En el fichero {} {}".format(user,
                                            sing_pl (lexpected - luser,
                                                     "falta la última línea.",
                                                     "faltan las {} últimas líneas.")
                                            )
    elif luser > lexpected:
        return "En el fichero {} hay {} de más.".format(user,
                                                        sing_pl(luser - lexpected,
                                                                "una línea",
                                                                "{} líneas")
                                                        )
    return None

def compararMatrices(expected, user):
    filasExpected, colsExpected = len(expected), len(expected[0])
    filasUser, colsUser = len(user), len(user[0])
    if filasExpected != filasUser or colsExpected != colsUser:
        return True, None
    for fila in range(len(expected)):
        for col in range(len(expected[0])):
            if expected[fila][col] != user[fila][col]:
                return ("En la fila {} y columna {} se ha encontrado el valor {} y se esperaba {}."
                         .format(fila, col, user[fila][col], expected[fila][col]))
    return None

def comprobarMatriz(image, result, conf):
    try:
        with open(image) as f:
            expected = matriz(f.read())
    except:
        return "No he podido abrir el fichero de pruebas {}, comprueba que existe.".format(image)
    user = result.globals["__builtins__"].get(conf.MATRIX_UJI)
    if user == None:
        return "No se ha mostrado ninguna matriz."
    return compararMatrices(expected, user)

def posDiferencia(cad1, cad2):
    if len(cad1) > len(cad2):
        cad1, cad2 = cad2, cad1
    for i in range(len(cad1)):
        if cad1[i] != cad2[i]:
            return i
    return len(cad1)

def prettyPrintDiferencias(encontrado, esperado):
    if encontrado != None and esperado != None:
        print("    Se esperaba: {0}".format(esperado))
        print("    Encontrado : {0}".format(encontrado))
        print(" "*(posDiferencia(esperado,encontrado)+15)+"**^**")
        print()
    elif esperado!= None:
        print("La salida no está completa. {}".format(
                           sing_pl(len(esperado), "Falta esta línea:",
                                   "Faltan las siguientes líneas:")))
        print(esperado)
        print()
    else:
        print("Hay más líneas de las esperadas. {}".format(
                           sing_pl(len(encontrado), "Sobra esta línea:",
                                   "Sobran las siguientes líneas:")))
        print(encontrado)
        print()

def redirectIO (input, output, error):
    stdin, stdout, stderr = sys.stdin, sys.stdout, sys.stderr
    sys.stdin, sys.stdout, sys.stderr = input, output, error
    return (stdin, stdout, stderr)

def restoreIO (io):
    sys.stdin.close()
    sys.stdout.close()
    sys.stderr.close()
    sys.stdin, sys.stdout, sys.stderr = io

class timeout:
    def __init__(self, seconds=1, error_message='Timeout'):
        self.seconds = seconds
        self.error_message = error_message

    def handle_timeout(self, signum, frame):
        raise TimeoutError(self.error_message)

    def __enter__(self):
        if not self.seconds is None:
            signal.signal(signal.SIGALRM, self.handle_timeout)
            signal.alarm(self.seconds)

    def __exit__(self, type, value, traceback):
        if not self.seconds is None:
            signal.alarm(0)

class executionResult:
    def __init__(self, value = None, output = None, error = None, exception = None, globals = None):
        self.value = value
        self.output = output
        self.error = error
        self.exception = exception
        self.globals = globals

class executionManager:
    def __init__(self, filename, timeout = None):
        self.timeout = timeout
        source = open(filename, encoding="utf-8").read()
        self.filename = filename
        result = self.do_exec(compile, (source, filename, "exec"), "")
        if result.error != "":
            self.error = result.error
            self.isOk = False
            return
        self.isOk = True
        self.prg = result.value
        tree = ast.parse(source, filename, "exec")
        tree.body = [ node for node in tree.body if node.__class__.__name__
                    in [ "FunctionDef", "ImportFrom", "Import", "ClassDef", "AsyncFunctionDef"] ]
        declarations = compile(tree, filename, "exec")
        self.globals = {}
        result = self.do_exec(exec, (declarations, self.globals), "")

    def exec_program(self, input):
        globals = {}
        result = self.do_exec(exec, (self.prg, globals), input)
        result.globals = globals
        return result

    def exists_function(self, fname):
        return fname in self.globals

    def exec_function(self, fname, pars, input):
        f = self.globals[fname]
        result = self.do_exec(f, pars, input)
        result.globals = self.globals
        return result

    def do_exec(self, f, pars, input):
        output = StringIO("")
        error = StringIO("")
        io = redirectIO(StringIO(input), output, error)
        exception = None
        value = None
        try:
            with timeout(self.timeout):
                value = f(*pars)
        except SystemExit:
            pass
        except Exception as e:
            exception = e
            lines = traceback.format_exception(*sys.exc_info())
            i = 0
            while i < len(lines):
                l = lines[i]
                if 'validador_old.py' not in l:
                    error.write(l)
                i += 1
        finally:
            result = executionResult(value = value, output = output.getvalue(), error = error.getvalue(), exception = exception)
            restoreIO(io)
        return result

def esta_implementado(fichero):
    return os.path.isfile(fichero) and os.stat(fichero).st_size > 0

def validacion(conf):
    resultado = Resultado()
    for ejercicio in conf.ejercicios:
        prueba_ejercicio(conf, ejercicio, resultado)
    return resultado

def prueba_ejercicio(conf, ejercicio, resultado):
    filename = ejercicio.nombre
    if ejercicio.obligatorio:
        resultado.add_to_zip(filename)
    if not esta_implementado(filename):
        resultado.add_not_implemented(filename, ejercicio.obligatorio)
        return
    if ejercicio.pruebas == []:
        print("{} no tiene pruebas (compruébalo manualmente)".format(filename))
        return
    isOk = True
    em = executionManager(filename, conf.TIMEOUT)
    if em.isOk:
        for prueba in ejercicio.pruebas:
            isOk = prueba.do_test(filename, em, conf)
            if not isOk:
                break
    else:
        isOk = False
        print ("{} no puede compilarse. Error:\n{}".format(filename, em.error))
    if isOk:
        print("{} pasa las pruebas".format(filename))
    else:
        resultado.add_not_valid(filename, ejercicio.obligatorio)

def get_desktop_folder():
    from os import path
    desktop = path.expanduser("~/Desktop")
    home = path.expanduser("~")
    if path.isdir(desktop):
        return desktop
    return home

class Resultado:
    def __init__(self):
        self.not_implemented_exercices = []
        self.not_implemented_mandatory_exercices = []
        self.not_valid_exercices = []
        self.not_valid_mandatory_exercices = []
        self.todas_las_pruebas_superadas = True
        self.todas_las_obligatorias_superadas = True
        self.to_be_zipped = []

    def add_to_zip(self, name):
        self.to_be_zipped.append(name)

    def add_not_implemented(self, name, mandatory):
        self.not_implemented_exercices.append(name)
        if mandatory:
            self.not_implemented_mandatory_exercices.append(name)
        self.add_not_valid(name, mandatory)

    def add_not_valid(self, name, mandatory):
        self.not_valid_exercices.append(name)
        self.todas_las_pruebas_superadas = False
        if mandatory:
            self.not_valid_mandatory_exercices.append(name)
            self.todas_las_obligatorias_superadas = False


class Configuración:
    fields = [("VERSION", "1.3.8"),
              ("TIMEOUT", 5), #seconds
              ("ENVIRONMENT_FLAG", "__MATRIX_GLOBAL__"),
              ("MATRIX_UJI", "__uji_matrix"),
              ("NUM_PRACTICA", -1),
              ("work", []),
              ("CREATE_ZIP", True),
              ("NUM_MAX_EJERCICIOS_MAL", 0) # de los obligatorios
              ]

    def __init__(self, variables):
        self.variables = variables
        for (nombre, valor) in Configuración.fields:
            self.asigna(nombre, valor)
        try:
            _ = signal.SIGALRM
            _ = TimeoutError
        except:
            self.TIMEOUT = None

    def asigna(self, nombre, valor):
        v = self.variables.get(nombre, valor)
        if nombre != "work":
            setattr(self, nombre, v)
        else:
            self.ejercicios = [ Ejercicio(e, n) for n, e in enumerate(v) ]


def has_len_and_items(what, description):
    if not hasattr(what, "__getitem__"):
        error(description + " no es del tipo adecuado, no tiene __getitem__")
    if not hasattr(what, "__len__"):
        error(description + " no es del tipo adecuado, no tiene __len__")

class Ejercicio:
    def __init__(self, entrada, n):
        desc = "la entrada {} de la lista work".format(n)
        has_len_and_items(entrada, desc)
        try:
            self.nombre = entrada[0]
        except:
            error("No puedo leer el primer elemento de {}".format(desc))
        self.pruebas = entrada[-1]
        self.obligatorio = len(entrada) == 2 or entrada[1]

class ProgramTest:
    def __init__(self, input, output, image = None, outputFile = None, refFile = None, functions = None):
        self.input = input
        self.output = output
        self.image = image
        self.outputFile = outputFile
        self.refFile = refFile
        self.functions = functions

    def do_test(self, filename, em, conf):
        if self.functions != None:
            isOk = self.check_functions(filename, conf, em)
            if not isOk:
                return False
        result = em.exec_program(self.input)
        if result.exception != None:
            if not conf.TIMEOUT is None and isinstance (result.exception, TimeoutError):
                print("{0} lanza un TIMEOUT (tarda mucho en ejecutarse) para entrada {1}.".format(filename, self.input.split()))
            else:
                print("{0} FALLO para entrada {1}. Lanzada excepción:".format(filename, self.input.split()))
                print(result.error)
            return False
        if result.error != "":
            print("{0} FALLO para entrada {1}. Salida de error:".format(filename, self.input.split()))
            print(result.error)
            return False
        return self.check_output(filename, result, conf)

    def check_output(self, filename, result, conf):
        if len(self.input)>0 and self.input[-1]=="\n":
            entrada = self.input[:-1].split("\n")
        else:
            entrada = self.input.split("\n")
        hayDiferencias, encontrado, esperado = comparaSalida(result.output, self.output)
        if hayDiferencias:
            print("{0} FALLO para entrada {1}.".format(filename, entrada))
            prettyPrintDiferencias(encontrado, esperado)
            return False
        ficherosDistintos = False
        if self.image != None:
            mensaje = comprobarMatriz(self.image, result, conf)
            if mensaje != None:
                print("{} FALLO para entrada {}. {}".format(filename, entrada, mensaje))
                return False
        if self.outputFile != None:
            mensaje = compararFicheros(self.refFile, self.outputFile)
            if mensaje != None:
                print("{} FALLO para entrada {}. {}".format(filename, entrada, mensaje))
                return False
        return True

    def check_functions(self, filename, conf, em):
        original = {}
        for k, v in em.globals.items():
            if k != '__builtins__':
                original[k] = my_deepcopy(v)
            else:
                original[k] = v

        for pf in self.functions:
            if not pf.do_test(filename, em, original):
                return False
        return True

class FunctionTestList:
    def __init__(self, fname, tests):
        self.fname = fname
        self.tests = tests

    def do_test(self, filename, em, original):
        if not em.exists_function(self.fname):
            print ("{} FALLO, no implementa la función {}.".format(filename, self.fname))
            return False
        for test in self.tests:
            if not test.do_test(self.fname, filename, em, original):
                return False
        return True

class FunctionTest:
    def __init__(self, pars, stdin, result, finalPars, stdout):
        self.pars = pars
        self.stdin = stdin
        self.result = result
        self.finalPars = finalPars
        self.stdout = stdout

    def do_test(self, fname, filename, em, original):
        parsActual = my_deepcopy(self.pars)
        result = em.exec_function(fname, parsActual, self.stdin)
        full = "{} FALLO, la función {} con parámetros {} y entrada {}".format(filename, fname, self.pars, repr(self.stdin))
        if result.exception != None:
            if isinstance (result.exception, TimeoutError):
                print("{} provoca un TIMEOUT (tarda mucho tiempo en ejecutarse).".format(full))
            else:
                print ("{} lanza una excepción:".format(full))
                print (result.error)
            return False
        if result.value != self.result:
            print ("{} da como resultado {} en lugar de {}".format(full, result.value, self.result))
            return False
        for (var, valor) in result.globals.items():
            if var not in original or valor != original[var]:
                print ("{} asigna a la variable global {}.".format(full, var))
                return False
        if self.finalPars != parsActual:
            print ("{} no trata los parametros como se espera.".format(full))
            for i,p in enumerate(self.finalPars):
                if p != parsActual[i]:
                    print("Al salir, el parámetro {} tenía que valer {} y vale {}".format(i + 1, p, parsActual[i]))
            return False
        hayDiferencias, encontrado, esperado = comparaSalida(result.output, self.stdout)
        if hayDiferencias:
            print("{} no da la salida correcta.".format(full))
            prettyPrintDiferencias(encontrado, esperado)
            return False
        return True

class ObjectTest:
    def __init__(self, oname, pars, tests):
        self.oname = oname
        self.pars = pars
        self.tests = tests

    def do_test(self, filename, em, conf):
        full = "{} FALLO, el constructor de {} con parámetros {}".format(filename, self.oname, self.pars)

        original = {}
        for k, v in em.globals.items():
            if k != "__builtins__":
                original[k] = my_deepcopy(v)
            else:
                original[k] = v

        parsActual = my_deepcopy(self.pars)
        if not em.exists_function(self.oname):
            print ("{} no existe la clase {}".format(filename, self.oname))
            return False

        result = em.exec_function(self.oname, parsActual, "")

        if result.exception != None:
            print ("{} lanza una excepción:".format(full))
            print (result.error)
            return False
        for (var, valor) in result.globals.items():
            if var not in original or valor != original[var]:
                print ("{} asigna a la variable global {}.".format(full, var))
                return False
        if self.pars != parsActual:
            print ("{} no trata los parametros como se espera.".format(full))
            for i,p in enumerate(self.pars):
                if p != parsActual[i]:
                    print("Al salir, el parámetro {} tenía que valer {} y vale {}".format(i + 1, repr(p), repr(parsActual[i])))
            return False
        if result.output != "":
            print("{} escribe en la salida estándar y no debería.".format(full))
            return False

        history = ["{}{}".format(self.oname, self.pars)]
        obj = result.value
        for mt in self.tests:
            if not mt.do_test(filename, em, conf, obj, original, history):
                return False
        return True

class MethodTest:
    def __init__(self, mname, pars, stdin, result, finalPars, stdout):
        self.mname = mname
        self.pars = pars
        self.stdin = stdin
        self.result = result
        self.finalPars = finalPars
        self.stdout = stdout

    def do_test(self, filename, em, conf, obj, original, history):
        parsActual = my_deepcopy(self.pars)
        clase = obj.__class__.__name__
        if not hasattr(obj, self.mname):
            print("{} la clase {} no tiene método {}".format(filename, clase, self.mname))
            return False
        history.append("{}{}".format(self.mname, self.pars))
        full = "{} FALLO, la clase {} tras hacer {}\n    {}\ny con entrada {}".format(
            filename, clase, sing_pl(len(history), "la llamada", "las llamadas"),
            "\n    ".join(history), repr(self.stdin))
        result = em.do_exec(getattr(obj, self.mname), parsActual, self.stdin)
        if result.exception != None:
            print ("{} lanza una excepción:".format(full))
            print (result.error)
            return False
        if result.value != self.result:
            print ("{} da como resultado {} en lugar de {}".format(full, repr(result.value), repr(self.result)))
            return False
        for (var, valor) in em.globals.items():
            if var not in original or valor != original[var]:
                print ("{} asigna a la variable global {}.".format(full, var))
                return False
        if self.finalPars != parsActual:
            print ("{} no trata los parametros como se espera.".format(full))
            for i,p in enumerate(self.finalPars):
                if p != parsActual[i]:
                    print("Al salir, el parámetro {} tenía que valer {} y vale {}".format(i + 1, repr(p), repr(parsActual[i])))
            return False
        hayDiferencias, encontrado, esperado = comparaSalida(result.output, self.stdout)
        if hayDiferencias:
            print("{} no da la salida correcta.".format(full))
            prettyPrintDiferencias(encontrado, esperado)
            return False
        return True

def lee_configuración():
    # busca el fichero de configuración del validador
    matching_files = glob.glob('./validador_prac[0-9]*.cfg')
    if len(matching_files) == 0:
        error("Falta el fichero de configuración del validador (validador_prac<num>.cfg). Copia a este directorio el que corresponda a esta práctica")
    elif len(matching_files)>1:
        error("Hay más de un fichero de configuración del validador (validador_prac<num>.cfg) en el directorio. Deja sólo el adecuado para esta práctica.")

    # carga el fichero de validación
    conf_file = matching_files[0]
    try:
        conf = open(conf_file, 'r', encoding="utf-8").read()
    except Exception as e:
        error("Fichero '{0}' no encontrado o no pudo abrirse".format(conf_file))

    variables = { "MANDATORY": True,
                  "OPTIONAL": False,
                  "ObjectTest" : ObjectTest,
                  "MethodTest" : MethodTest,
                  "ProgramTest" : ProgramTest,
                  "FunctionTest" : FunctionTest,
                  "FunctionTestList" : FunctionTestList
            }

    try:
        exec(conf, variables)
    except Exception as e:
        error ("Formato erróneo en el fichero de configuración '{}'.\nExcepción:\n{}".format(conf_file, e))

    conf = Configuración(variables)

    # comprueba 'work'
    if len(conf.ejercicios)==0:
        error ("El fichero de configuración '{0}' no tiene ninguna prueba.".format(matching_files[0]))
    return conf

def crea_zip(conf, resultado):
    print("   GENERANDO ZIP DE ENTREGA...")

    ZIP_ENTREGA = "entrega_prac{0}.zip".format(conf.NUM_PRACTICA)
    carpeta_destino = get_desktop_folder()
    filename = carpeta_destino+"/"+ZIP_ENTREGA
    zf = zipfile.ZipFile(filename, mode='w')
    try:
        for ejfname in resultado.to_be_zipped:
            if not os.path.isfile(ejfname):
                print("\t\t->ERROR: el fichero '{0}' no existe. Para que el zip se genere, DEBES crearlo, aunque esté vacío.".format(ejfname))
        for ejfname in resultado.to_be_zipped:
            zf.write(ejfname)
        print("   ...ZIP DE ENTREGA GENERADO EN:")
        print("\n\t{0}".format(filename))
    except:
        print("...ERROR validador: No se pudo generar el zip de entrega.")
    finally:
        zf.close()

def cabecera(conf):
    msg1 = "VALIDANDO PRÁCTICA {0}".format(conf.NUM_PRACTICA)
    msg2 = "VALIDADOR v{0}".format(conf.VERSION)
    msg3 = "[ejecutándose en Python {0}.{1}.{2}]".format(version_info.major, version_info.minor, version_info.micro)
    ll = max(len(msg1),len(msg2),len(msg3))
    print("-"*ll)
    print(msg1)
    print("-"*ll)
    print(msg2)
    print(msg3)
    print("-"*ll)


def opciones():
    while True:
        print("""Opciones:
    + Pulsa INTRO para validar todo y generar el fichero de entrega.
    + Introduce un nombre de fichero para validarlo (p.ej. ej03.py)
    + Introduce un número de fichero sin .py para validarlo con .py (p.ej. ej03 para validar ej03.py)
    + Introduce un número para validar ese ejercicio (p.ej. 3 para validar ej03.py)
    + Introduce g para generar los ficheros vacíos que falten.
    + Introduce x para salir sin hacer nada.
    """)
        opción = input("¿Qué eliges? ")
        if opción == "":
            return []
        elif opción == "x":
            sys.exit()
        elif opción == "g":
            return opción
        else:
            if opción.startswith("ej"):
                if opción.endswith(".py"):
                    nombre = opción
                else:
                    nombre = opción + ".py"
            else:
                try:
                    n = int(opción)
                except ValueError:
                    print("Creo que has escrito mal el nombre del ejercicio, prueba otra vez")
                    continue
                nombre = "ej{:02}.py".format(n)
            if not esta_implementado(nombre):
                print ("El ejercicio {} no está implementado, prueba otra opción.".format(nombre))
            else:
                return [nombre]

def genera_ficheros(conf):
    for ejercicio in conf.ejercicios:
        if not os.path.isfile(ejercicio.nombre):
            open(ejercicio.nombre, "w").close()

def valida_uno(conf, nombre):
    print("\nProbando:")
    os.environ[conf.ENVIRONMENT_FLAG] = 'TRUE'
    resultado = Resultado()
    for ejercicio in conf.ejercicios:
        if ejercicio.nombre == nombre:
            prueba_ejercicio(conf, ejercicio, resultado)
            return

def valida_todos(conf):
    os.environ[conf.ENVIRONMENT_FLAG] = 'TRUE'
    resultado = validacion(conf)

    # Muestra la lista de ejercicios vacíos o que no existe el fichero
    num_not_implemented = len(resultado.not_implemented_exercices)
    if num_not_implemented > 0:
        num_mandatory_nie = len(resultado.not_implemented_mandatory_exercices)
        num_optional_nie = num_not_implemented - num_mandatory_nie
        print("\nEJERCICIOS NO IMPLEMENTADOS (total: {0}): ".format(num_not_implemented), end="")
        if num_mandatory_nie == 0:
            print("todos opcionales\n")
        else:
            print("{}{}\n".format(' '.join(str(f) for f in resultado.not_implemented_mandatory_exercices),
                  " (y opcionales: {})".format(num_optional_nie) if num_optional_nie > 0 else ""))

    print("<VALIDACIÓN FINALIZADA>")
    print()

    # Muestra el resultado final y, si corresponde, genera zip de entrega
    if resultado.todas_las_obligatorias_superadas or len(resultado.not_valid_mandatory_exercices) <= conf.NUM_MAX_EJERCICIOS_MAL:
        print('-'*20)
        print()
        if resultado.todas_las_pruebas_superadas:
            print("RESULTADO FINAL: VALIDACIÓN COMPLETAMENTE SUPERADA")
        elif resultado.todas_las_obligatorias_superadas:
            print("RESULTADO FINAL: VALIDACIÓN SUPERADA POR OBLIGATORIAS")
        elif len(resultado.not_valid_mandatory_exercices)==1:
            print("RESULTADO FINAL: VALIDACIÓN SUPERADA POR OBLIGATORIAS, TOLERANDO un ejercicio ({0}) que NO supera las pruebas".format(','.join(resultado.not_valid_mandatory_exercices)))
        else:
            print("RESULTADO FINAL: VALIDACIÓN SUPERADA POR OBLIGATORIAS, TOLERANDO {0} ejercicios ({1}) que NO superan las pruebas".format(len(resultado.not_valid_mandatory_exercices),
                                                                                                                                            ','.join(resultado.not_valid_mandatory_exercices)))
        print()
        print('-'*20)
                                                                                                                                            
        if conf.CREATE_ZIP:
            crea_zip(conf, resultado)
    # Septiembre 2024. En caso de error, no muestra ningún mensaje
    #else:
    #    print("RESULTADO FINAL: VALIDACIÓN NO SUPERADA")

def main():
    conf = lee_configuración()
    # lanza la validación
    cabecera(conf)
    op = opciones()
    if op == "g":
        genera_ficheros(conf)
    elif op == []:
        valida_todos(conf)
    else:
        valida_uno(conf, op[0])

if __name__ == "__main__":
    if version_info < (3, 1):
        error("El validador sólo funciona sobre Python 3.1 o superior")
    main()
