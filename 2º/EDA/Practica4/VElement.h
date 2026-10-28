/**
 * Practicas de EDA.
 * Curso 2012-2013
 * Departament d'Informatica.
 * Universitat de Valencia.
 */


#ifndef _VELEMENT_H
#define _VELEMENT_H


/**
 * Clase de elementos de vector. Los elementos de vector son enteros.
 * Esta clase ofrece funciones para el manejo de elementos de vector.
 * Mediante el uso de estas funciones la clase lleva la cuenta en variables
 * estáticas de las comparaciones y asignaciones que un algoritmo determinado realiza.
 *
 */
class VElement
{
    private:
      
        // Valor del elemento del vector
        int value;

        // Contadores de asignaciones y comparaciones
        static int nAssign;
        static int nComparison;

    public:
     
        /**
        * Contructor por defecto crea un VElement con valor 0.
        */
        VElement();

        /**
        * Contructor de copia, crea un VElement con valor v.
        *
        * @param v valor entero de inicialización.
        */
        VElement(int v);

        /**
        * Contructor de copia, crea un VElement con valor vE.
        *
        * @param vE elemento de inicialización.
        */
        VElement(const VElement &vE);

        /**
        * Cambia el valor del elemento. NO incrementa contadores.
        *
        * @param v nuevo valor del elemento.
        */
        void set(int v);

        /**
        * Retorna el valor del elemento. NO incrementa contadores.
        *
        * @return valor del elemento.
        */
        int get() const;

        /**
        * Sobrecarga del operador mayor.
        *
        * @param vE VElement con el que se compara valor.
        */
        bool operator >(const VElement &vE) const;

        /**
        * Sobrecarga del operador mayor.
        *
        * @param e valor con el que se compara valor.
        */
        bool operator >(int e) const;

        /**
        * Sobrecarga del operador menor.
        *
        * @param vE VElement con el que se compara valor.
        */
        bool operator <(const VElement &vE) const;

        /**
        * Sobrecarga del operador menor.
        *
        * @param e valor con el que se compara valor.
        */
        bool operator <(int e) const;

        /**
        * Sobrecarga del operador mayor o igual.
        *
        * @param vE VElement con el que se compara valor.
        */
        bool operator >=(const VElement &vE) const;

        /**
        * Sobrecarga del operador mayor o igual.
        *
        * @param e valor con el que se compara valor.
        */
        bool operator >=(int e) const;

        /**
        * Sobrecarga del operador menor o igual.
        *
        * @param vE VElement con el que se compara valor.
        */
        bool operator <=(const VElement &vE) const;

        /**
        * Sobrecarga del operador menor o igual.
        *
        * @param e valor con el que se compara valor.
        */
        bool operator <=(int e) const;

        /**
        * Sobrecarga del operador igual.
        *
        * @param vE VElement con el que se compara valor.
        */
        bool operator ==(const VElement &vE) const;

        /**
        * Sobrecarga del operador igual.
        *
        * @param e valor con el que se compara valor.
        */
        bool operator ==(int e) const;

        /**
        * Sobrecarga del operador asignación.
        *
        * @param e entero cuyo valor se asigna al VElement.
        */
        VElement & operator =(int e);

        /**
        * Sobrecarga del operador asignación.
        *
        * @param vE entero cuyo valor se asigna al VElement.
        */
        VElement & operator =(const VElement &vE);

        /**
        * Función que inicializa el contador de intervambios.
        *
        */
        static void initNAssign()
        {
            nAssign = 0;
        }

        /**
        * Retorna el número de intercambios.
        *
        * @return número de intercambios contadas.
        */
        static int getNAssign()
        {
            return nAssign;
        }

        /**
        * Función que inicializa el contador de comparaciones.
        *
        */
        static void initNComparison()
        {
            nComparison = 0;
        }

        /**
        * Retorna el número de comparaciones.
        *
        * @return número de comparaciones contadas.
        */
        static int getNComparison()
        {
            return nComparison;
        }
};


#endif
