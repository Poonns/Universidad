/**
 * @file costes.cpp
 * @brief Cálculo de costes teóricos y reales de funciones fa y fb con representación gráfica.
 *
 * Práctica 2 - Estructuras de Datos y Algoritmos
 * Cálculo de coste mejor, peor y medio (función fb)
 */

#include <iostream>
#include <fstream>
#include <cstdlib>
#include <ctime>
using namespace std;

const int MAX = 100;      ///< Tamaño máximo del vector
typedef int Vector[MAX]; ///< Alias para un vector de enteros

/* ================================================================
 *             PROTOTIPOS - DOCUMENTADOS CON DOXYGEN
 * ================================================================*/

/**
 * @brief Función fa del enunciado, mide su coste en pasos.
 * @param n Tamaño del problema
 * @param cont Contador de pasos (por referencia)
 * @return Resultado calculado (no relevante para el coste)
 */
unsigned int fa(unsigned int n, int &cont);

/**
 * @brief Función fb del enunciado (comprueba si el vector es cíclico)
 * @param v Vector de entrada con valores {0,1}
 * @param n Tamaño del vector
 * @param cont Contador de pasos (por referencia)
 * @return @c true si es cíclico, @c false en caso contrario
 */
bool fb(Vector v, unsigned int n, int &cont);

/**
 * @brief Genera un vector aleatorio de 0 y 1.
 * @param v Vector que se rellenará
 * @param n Tamaño del vector
 */
void generarVectorAleatorio(Vector v, unsigned int n);

/* ---- Funciones auxiliares coste real ---- */

/**
 * @brief Devuelve el nº de pasos de fa para tamaño @a n.
 */
float faAux(unsigned int n);

/**
 * @brief Devuelve pasos de fb en mejor caso.
 */
float fbAuxMejor(unsigned int n);

/**
 * @brief Devuelve pasos de fb en peor caso.
 */
float fbAuxPeor(unsigned int n);

/**
 * @brief Devuelve coste medio REAL de fb usando medias estadísticas.
 */
float fbAuxMedio(unsigned int n);

/* ---- Funciones auxiliares coste teórico ---- */

/**
 * @brief Coste teórico de fa(n) = 2n²
 */
float faTeo(unsigned int n);

/**
 * @brief Coste teórico mejor caso fb = 3n + 4
 */
float fbTeoMejor(unsigned int n);

/**
 * @brief Coste teórico peor caso fb = 3n + 7
 */
float fbTeoPeor(unsigned int n);

/**
 * @brief Coste teórico medio fb = (3n)/2 + 11/2
 */
float fbTeoMedio(unsigned int n);

/**
 * @brief Evalúa una función de coste para varios valores y los vuelca a fichero
 *
 * @param ini Tamaño inicial
 * @param fin Tamaño final
 * @param inc Incremento
 * @param t Función de coste
 * @param file Flujo de salida
 */
void evaluar(unsigned int ini, unsigned int fin, int inc,
             float (*t)(unsigned), ostream &file);


/* ================================================================
 *                            MAIN
 * ================================================================*/

/**
 * @brief Genera los ficheros .dat con los costes de fa y fb
 */
int main() {
    srand(time(NULL));

    ofstream at_real("at_real.dat");
    ofstream at_teo("at_teo.dat");

    ofstream btm_real("btm_real.dat");
    ofstream btm_teo("btm_teo.dat");

    ofstream btp_real("btp_real.dat");
    ofstream btp_teo("btp_teo.dat");

    ofstream btmid_real("btmid_real.dat");
    ofstream btmid_teo("btmid_teo.dat");

    evaluar(4, 100, 4, faAux, at_real);
    evaluar(4, 100, 4, faTeo, at_teo);

    evaluar(4, 100, 4, fbAuxMejor, btm_real);
    evaluar(4, 100, 4, fbTeoMejor, btm_teo);

    evaluar(4, 100, 4, fbAuxPeor, btp_real);
    evaluar(4, 100, 4, fbTeoPeor, btp_teo);

    evaluar(4, 100, 4, fbAuxMedio, btmid_real);
    evaluar(4, 100, 4, fbTeoMedio, btmid_teo);

    return 0;
}

/* ================================================================
 *                            FUNCIONES
 * ================================================================*/

unsigned int fa(unsigned int n, int &cont) {
    cont = 1;
    unsigned int k = 0;
    for (unsigned int i = n; i >= 2; i--) {
        cont++;
        for (unsigned int j = i; j <= n; j++) {
            cont++;
            k = k + j;
        }
    }
    cont++;
    return k;
}

bool fb(Vector v, unsigned int n, int &cont) {
    cont = 0;
    int a = 0; cont++;
    int b = n - 1; cont++;

    while ((a < b) && (v[a] == v[b])) {
        cont++;
        a++; cont++;
        b--; cont++;
    }
    cont++;

    bool cic = (a >= b);
    cont++;

    return cic;
}

void generarVectorAleatorio(Vector v, unsigned int n) {
    for(unsigned int i = 0; i < n; i++)
        v[i] = rand() % 2;
}

/* ==================== COSTE REAL ==================== */

float faAux(unsigned int n) {
    int pasos;
    fa(n, pasos);
    return pasos;
}

float fbAuxMejor(unsigned int n) {
    int pasos;
    Vector v;
    for(unsigned int i = 0; i < n; i++) v[i] = 1;
    fb(v, n, pasos);
    return pasos;
}

float fbAuxPeor(unsigned int n) {
    int pasos;
    Vector v;
    for(unsigned int i = 0; i < n; i++) v[i] = 0;
    v[n-1] = 1;
    fb(v, n, pasos);
    return pasos;
}

float fbAuxMedio(unsigned int n) {
    const int REP = 1000;  ///< nº de repeticiones para media estadística
    int pasos;
    float total = 0;
    Vector v;

    for(int i = 0; i < REP; i++) {
        generarVectorAleatorio(v, n);
        fb(v, n, pasos);
        total += pasos;
    }
    return total / REP;
}

/* ==================== COSTE TEÓRICO ==================== */

float faTeo(unsigned int n) {
    return 2*n*n;
}

float fbTeoMejor(unsigned int n) {
    return 3*n + 4;
}

float fbTeoPeor(unsigned int n) {
    return 3*n + 7;
}

float fbTeoMedio(unsigned int n) {
    return (3.0*n)/2.0 + 11.0/2.0;
}

/* ==================== EVALUACIÓN ==================== */

void evaluar(unsigned int ini, unsigned int fin, int inc,
             float (*t)(unsigned), ostream &file) {
    for(unsigned int i = ini; i <= fin; i += inc)
        file << i << " " << t(i) << endl;
}


