package Hector.Prac3;

public class ListaCadenasEnlaceSimple {
    int talla;
    ListaCadenasEnlaceSimple sig;
    String dato;

    public interface ListaCadenas {
        boolean add(String s);
        void add(int i, String s);
        void clear();
        String get(int i);
        int indexOf(String s);
        int lastIndexOf(String s);
        boolean isEmpty();
        String remove(int i);
        String set(int i, String s);
        int size();
        }

    public boolean add(String s){
        ListaCadenasEnlaceSimple nuevo = new ListaCadenasEnlaceSimple();
        nuevo.dato = s;
        nuevo.sig = null;

        if(sig == null){
            sig = nuevo;
        } else{
            for(ListaCadenasEnlaceSimple actual = sig; actual != null; actual = actual.sig){
                if(actual.sig == null){
                    actual.sig = nuevo;
                    break;
                }
            }
        }
        talla++;
        return true;
    }

    public void add(int i, String s) throws IndexOutOfBoundsException{
        ListaCadenasEnlaceSimple nuevo = new ListaCadenasEnlaceSimple();
        nuevo.dato = s;
        nuevo.sig = null;

        if(i == 0){
            nuevo.sig = sig;
            sig = nuevo;
            talla++;
        }

        ListaCadenasEnlaceSimple actual = sig;
        for(int j = 0; j < i -1; j++){
            if(actual == null){
                throw new IndexOutOfBoundsException("Posicion fuera de rango");
            }
            actual = actual.sig;
        }
        nuevo.sig = actual.sig;
        actual.sig = nuevo;

        talla++;
    }

    public void clear(){
        for(ListaCadenasEnlaceSimple actual = sig; actual != null; actual = actual.sig){
            actual.sig = null;
        }
        sig = null;
        talla = 0;
    }

    public String get(int i) throws IndexOutOfBoundsException{
        ListaCadenasEnlaceSimple actual = sig;
        for(int j = 0; j < i; j++){
            if(actual == null){
                throw new IndexOutOfBoundsException("Posicion fuera de rango");
            }
            actual = actual.sig;
        }

        if(actual == null){
            throw new IndexOutOfBoundsException("Posicion fuera de rango");
        }
        return actual.dato;
    }

    public int indexOf(String s){
        int contador = 0;
        for(ListaCadenasEnlaceSimple actual = sig; actual != null; actual = actual.sig){
            if(actual.dato.equals(s)){
                return contador; 
            }
            contador++;
            
        }
        return -1;
    }

    public int lastIndexOf(String s){
        int contador = 0;
        for(ListaCadenasEnlaceSimple actual = sig; actual != null; actual = actual.sig){
            if(actual.dato.equals(s)){
                contador++;
            }
        }
        if(contador == 0){
            return -1;
        }
        return contador;
    }

    public boolean isEmpty(){
        if(sig == null){
            return true;
        } else {
            return false;
        }
    }

    public String remove(int i) throws IndexOutOfBoundsException{
        if(sig == null){
            throw new IndexOutOfBoundsException("Lista vacia");
        }
        
        String dato;

        if(i == 0){
            dato = sig.dato;
            sig = sig.sig;
            talla --;
        } else {
            ListaCadenasEnlaceSimple actual = sig;
            for(int j = 0; j < i; j++){
                if(actual.sig == null){
                    throw new IndexOutOfBoundsException("Posicion fuera de rango");
                }
                actual = actual.sig;
            }
            dato = actual.sig.dato;
            actual.sig = actual.sig.sig;
            talla --;
        }
        return dato;
    }

    public String set(int i, String s) throws IndexOutOfBoundsException{
        if(i < 0){
            throw new IndexOutOfBoundsException("Posicion no valida");
        }
        if(sig == null){
            throw new IndexOutOfBoundsException("Lista vacia");
        }
        ListaCadenasEnlaceSimple actual = sig;
        for(int j = 0; j < i; j++){
            if(actual == null){
                throw new IndexOutOfBoundsException("Posicion no valida");
            }
            actual = actual.sig;
        }
        if(actual == null){
            throw new IndexOutOfBoundsException("Posicion no valida");
        }
        actual.sig.dato = s;
        return actual.sig.dato;
    }
    
    public int size(){
        int contador = 0;
        for(ListaCadenasEnlaceSimple actual = sig; actual != null; actual = actual.sig){
            contador++;
        }
        return contador;
    }
    
    public String toString(){
        String cadena = "[";
        for(ListaCadenasEnlaceSimple actual = sig; actual != null; actual = actual.sig){
            cadena += actual.dato;
            if(actual.sig != null){
                cadena += ",";
            }
        }
        cadena += "]";
        return cadena;
    }
}
