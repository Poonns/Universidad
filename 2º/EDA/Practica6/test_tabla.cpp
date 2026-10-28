/**
 * @file test_tabla.cpp
 * @brief Programa de prueba (Test Suite) para la clase Tabla Hash.
 * @details Este programa verifica el funcionamiento correcto de la inserción, búsqueda,
 * cálculo del factor de carga y distribución de colisiones de la Tabla Hash.
 * @date Curso 2025/2026
 */

#include <iostream>
#include <vector>
#include <string>
#include <cstdlib>
#include <algorithm>
#include <ctime> 
#include "alumno.h"
#include "tabla.h" 

using namespace std;

/** @brief Número de elementos que se insertarán en la prueba. */
const unsigned TALLA = 1000;

/** @brief tamanyo de la tabla. */
const unsigned TAM_TABLA = 2000;


// PROTOTIPOS DE LAS FUNCIONES
unsigned obtenerClave(const string&);
void prueba();

/**
 * @brief Función principal.
 * Inicializa la semilla aleatoria y lanza la batería de pruebas.
 * @return 0 si la ejecución fue correcta.
 */
int main() 
{
    // Inicializamos la semilla del generador de números aleatorios con el tiempo actual
    // para que cada ejecución genere alumnos diferentes.
    srand(time(0)); 
    prueba();
    return 0;
}

/**
 * @brief Extrae la parte numérica de un DNI formateado.
 * @details El DNI suele tener formato "12345678A". Esta función elimina la letra final
 * para usar el número como clave en la Tabla Hash.
 * @param dni Cadena de texto con el DNI completo.
 * @return El número del DNI convertido a unsigned int.
 */
unsigned obtenerClave(const string& dni)
{
    // Tomamos la subcadena desde el inicio (0) hasta longitud-1 (quitamos la letra)
    string clave = dni.substr(0, dni.length() - 1);
    // stoul: String To Unsigned Long
    return stoul(clave);
}

/**
 * @brief Función que ejecuta la lógica de prueba de la Tabla Hash.
 * @details Realiza las siguientes operaciones:
 * 1. Inserta 1000 alumnos aleatorios.
 * 2. Muestra el estado de la tabla.
 * 3. Verifica búsquedas exitosas (elementos que sabemos que existen).
 * 4. Verifica búsquedas fallidas (elementos que no existen).
 * 5. Muestra estadísticas de ocupación (factor de carga y colisiones).
 */
void prueba() 
{
    // --- Declaración de variables ---
    unsigned clave_numerica, clave_existente_1, clave_existente_2, clave_no_existente;
    string dni_completo, dni_existente_1, dni_existente_2;
    
    // Instanciamos la tabla con el tamaño definido
    Tabla<unsigned, Alumno> tabla_alumnos(TAM_TABLA); 
    
    // Vector auxiliar para guardar los DNIs generados y poder buscarlos después para comprobar
    vector<string> dnis_insertados; 
    Alumno alu_encontrado;

    // --- 1. Fase de Inserción Masiva ---
    cout << "Generando e insertando " << TALLA << " alumnos..." << endl;
    for (unsigned i = 0; i < TALLA; ++i) 
    {
        // Creamos un alumno con datos aleatorios (el constructor con int suele generar random)
        Alumno alu(1);
        
        // Obtenemos su DNI y extraemos la clave numérica
        dni_completo = alu.getDNI();
        clave_numerica = obtenerClave(dni_completo);
        
        // Insertamos en la tabla hash: Clave (unsigned) -> Valor (Objeto Alumno)
        tabla_alumnos.insertar(clave_numerica, alu);

        // Guardamos el DNI en nuestro vector auxiliar de control
        dnis_insertados.push_back(dni_completo); 
    }

    // --- 2. Visualización Parcial ---
    cout << "\n1. CONTENIDO DE LA TABLA (Muestra Parcial) " << endl;
    // Nota: Si TALLA es muy grande, esto imprimirá muchas líneas por consola.
    tabla_alumnos.mostrar(cout); 

    // --- 3. Pruebas de Búsqueda (Casos de Éxito) ---
    // Recuperamos el primer DNI que insertamos para asegurar que existe
    dni_existente_1 = dnis_insertados[0];
    clave_existente_1 = obtenerClave(dni_existente_1);
    
    cout << "\n2. BÚSQUEDAS EN LA TABLA " << endl;
    cout << "Buscando DNI EXISTENTE (Clave: " << clave_existente_1 << "):" << endl;
    
    // El método buscar devuelve true si encuentra, y rellena 'alu_encontrado' por referencia
    if (tabla_alumnos.buscar(clave_existente_1, alu_encontrado)) 
    {
        cout << "Alumno encontrado: " << endl;
        cout << alu_encontrado; 
    } 
    else 
    {
        cout << "ERROR: Alumno no encontrado." << endl;
    }
    
    // Probamos con otro elemento (el que está a mitad del vector auxiliar)
    dni_existente_2 = dnis_insertados[TALLA / 2];
    clave_existente_2 = obtenerClave(dni_existente_2);

    cout << "-> Buscando DNI EXISTENTE (Clave: " << clave_existente_2 << "):" << endl;
    if (tabla_alumnos.buscar(clave_existente_2, alu_encontrado)) 
    {
        cout << "Alumno encontrado: " << endl;
        cout << alu_encontrado;
    } 
    else 
    {
        cout << "ERROR: Alumno no encontrado" << endl;
    }
    
    // --- 4. Prueba de Búsqueda (Caso de Fallo) ---
    clave_no_existente = 999999999U; // Usamos un número muy alto improbable

    cout << "-> Buscando DNI NO EXISTENTE (Clave: " << clave_no_existente << "):" << endl;
    if (tabla_alumnos.buscar(clave_no_existente, alu_encontrado)) 
    {
        cout << "ERROR: Se ha encontrado una clave que no debería existir."<< endl;
    } 
    else 
    {
        cout << "Correcto: No encontrado." << endl;
    }


    // --- 5. Estadísticas de Rendimiento ---
    cout << "\n3. FACTOR DE CARGA Y DISTRIBUCIÓN " << endl;
    // El factor de carga idealmente debería ser cercano a 1.0 si TALLA == TAM_TABLA
    cout << "Factor de Carga (alfa): " << tabla_alumnos.factorCarga() << endl;
    
    // num_posN nos dice la calidad de la función hash. 
    // Si hay muchas posiciones con 0 elementos y otras con muchos (3, 4, etc.),
    // significa que hay muchas colisiones y la distribución no es uniforme.
    cout << "Posiciones con 0 elementos (Vacías): " << tabla_alumnos.num_posN(0) << endl;
    cout << "Posiciones con 1 elemento (Ideal):   " << tabla_alumnos.num_posN(1) << endl;
    cout << "Posiciones con 2 elementos (Colisión): " << tabla_alumnos.num_posN(2) << endl;
    cout << "Posiciones con 3 elementos (Colisión): " << tabla_alumnos.num_posN(3) << endl;
}